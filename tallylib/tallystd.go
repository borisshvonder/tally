package tallylib

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"strings"
)

// Holds settings
type tally struct {
	config      TallyConfig
	log         io.Writer
	collectionPathnameTemplate *template.Template
	loggerDebug *log.Logger
	loggerInfo  *log.Logger
	loggerErr   *log.Logger
	loggerWarn  *log.Logger
}

func NewTally() Tally {
	var ret = new(tally)
	ret.config.LogVerbosity = 3
	ret.config.CollectionPathnameExpression = "{{.Path 0}}.rscollection"
	ret.SetLog(ioutil.Discard)
	return ret
}

func (tally *tally) GetConfig() TallyConfig {
	return tally.config
}

func (tally *tally) SetConfig(cfg TallyConfig) {
	tally.config = cfg
	tally.collectionPathnameTemplate = nil
}

func (tally *tally) SetLog(logfile io.Writer) {
	tally.log = logfile
	tally.initLog()
}

func (tally *tally) init(directory string) (string, error)  {
	var err = tally.ensureTemplateCompiled()
	var ret string
	if err == nil {
		ret = filepath.Clean(directory)
	}
	return ret, err
}

func (tally *tally) UpdateRecursive(directory string) (bool, error) {
	var normalizedPath, err = tally.init(directory)
	if err != nil {
		return false, err
	}
	tally.info("UpdateRecursive(", normalizedPath, ")")
	err = tally.assertDirectory(normalizedPath)
	if err != nil {
		return false, err
	}
	
	tally.debug("Stage1: updating children")
	var ret bool
	
	ret, err = tally.updateChildren(normalizedPath)
	if err != nil {
		return ret, err
	}


	if tally.config.UpdateParents && ret {
		tally.debug("Stage3: updating parents")
		_, err = tally.updateParents(normalizedPath)
	}

	return ret, err
}

func (tally *tally) updateChildren(directory string) (bool, error) {
	var files, err = tally.listDirectory(directory)

	var ret = false	
	var changed = false
	
	for _, file := range files {
		if tally.isDir(file) {
			tally.debug("Invoking updateChildren(", file.Name(), ")")
			var fullpath = filepath.Join(directory, file.Name())
			changed, err = tally.updateChildren(fullpath)
			ret = ret || changed
			if err != nil {
				return ret, err
			}
		} else {
			tally.debug("Skipping file cause it is not directory", file.Name())
		}
	}

	changed, err = tally.UpdateSingleDirectory(directory)
	ret = ret || changed
	if err != nil {
		return ret, err
	}

	return ret, err
}

func (tally *tally) updateParents(directory string) (bool, error) {
	var ret = false
	var err error
	
	for parent := filepath.Dir(directory); err == nil && parent != "/"; parent = filepath.Dir(parent) {
		var collectionFile string 
		collectionFile, err = tally.resolveCollectionFileForDirectory(parent)
		if err != nil {
			return false, err
		}
		var stat os.FileInfo
		stat, err = os.Stat(collectionFile)
		if err == nil {
			if !tally.isFile(stat) {
				tally.info("Stopping updating patents, file", collectionFile, "is not a regualr file")
				break
			} else {
				tally.debug("file ", collectionFile, "found")
				tally.info("Updating parent", parent)

				var changed bool
				changed, err = tally.UpdateSingleDirectory(parent)
				ret = ret || changed
				if err != nil {
					return ret, err
				}
			}
		} else {
			if os.IsNotExist(err) {
				tally.info("Stopping updating parents, file", collectionFile, "not found")
			} else {
				tally.info("Stopping updating parents, stat error on file", collectionFile, ":", err)
			}
		}
	}

	return ret, nil
}

func (tally *tally) UpdateSingleDirectory(directory string) (bool, error) {
	var normalizedPath, err = tally.init(directory)
	if err != nil {
		return false, err
	}
	tally.info("UpdateSingleDirectory(", normalizedPath, ")")
	err = tally.assertDirectory(directory)
	if err != nil {
		return false, err
	}

	// In order to remove files which no longer exists from the collection,
	// we use 2 collections here: oldColl and newColl. This is actually
	// does not consume too much memory since collections store interfaces
	// (pointers) to real data. Regardless, the data is very small anyway
	var collectionFile string
	collectionFile, err = tally.resolveCollectionFileForDirectory(normalizedPath)
	if err != nil {
		return false, err
	}
	tally.debug("Using collection file ", collectionFile)
	var newColl, oldColl RSCollection
	oldColl, err = tally.loadExistingCollection(collectionFile)
	if err != nil {
		tally.debug("Error loading from ", collectionFile, err)
		return false, err
	}
	newColl = NewCollection()
	newColl.InitEmpty()

	var files []os.FileInfo
	files, err = tally.listDirectory(normalizedPath)
	var ret = false

	for _, file := range files {
		if tally.isFile(file) {
			var changed bool
			var name = file.Name()
			tally.debug("Working on file", name)

			var oldFile = oldColl.ByName(name)
			oldColl.RemoveFile(name)

			if oldFile != nil {
				newColl.UpdateFile(oldFile)
			}
			changed, err = tally.updateFile(normalizedPath, name, newColl)
			if err != nil {
				return ret, err
			}
			if changed {
				ret = true
				tally.info("Detected change in", name)
			}
		} else {
			tally.debug("Skipping", file.Name(), "because it is not a regular file")
		}
	}

	// files left in old collection could mean files were removed from disk
	if tally.config.RemoveExtraFiles {
		ret = ret || oldColl.Size() > 0
	} else {
		var changed = tally.removeMissingFiles(normalizedPath, oldColl, newColl)
		ret = ret || changed
	}

	if ret {
		// Collection has been modified, need to write it back
		err = tally.storeCollectionToFile(newColl, collectionFile)
	}

	return ret, err
}

func (tally *tally) listDirectory(directory string) ([]os.FileInfo, error) {
	tally.debug("Listing files in directory ", directory)
	var files, err = ioutil.ReadDir(directory)
	if err != nil {
		err = tally.accessError(directory, "Can't list", err)
		return nil, err
	}
	tally.debug("Got", len(files), "entries")
	return files, nil
}

func (tally *tally) assertDirectory(directory string) error {
	var stat, err = os.Stat(directory)
	if err != nil {
		if os.IsNotExist(err) {
			err = tally.accessError(directory, "does not exist", err)
		} else {
			err = tally.accessError(directory, "Can't stat", err)
		}
		tally.err(err)
		return err
	}
	if !tally.isDir(stat) {
		err = tally.accessError(directory, "supplied path is not a directory", nil)
		tally.err(err)
		return err
	}
	return nil
}

func (tally *tally) removeMissingFiles(directory string, oldColl, newColl RSCollection) bool {
	var ret = false
	oldColl.Visit( func(rsfile RSCollectionFile) {
		var fullpath = filepath.Join(directory, rsfile.Name())
		var _, err = os.Stat(fullpath)
		if err != nil && os.IsNotExist(err) {
			tally.info("File", fullpath, "has gone from disk, removing")
			ret = true
		} else {
			tally.debug("Keeping", rsfile.Name(), "in collection since I can't tell if it was removed")
			newColl.UpdateFile(rsfile)
		}
	})
	return ret
}

func (tally *tally) updateFile(directory, filename string, coll RSCollection) (bool, error) {
	tally.debug("Checking file", filename, "in directory", directory)
	var fullpath = filepath.Join(directory, filename)
	var ret, err = updateFile(coll, filename, fullpath, tally.config.ForceUpdate)

	if err != nil {
		// Failure to update single file is not critical
		tally.warn("Could not update", fullpath, err)
		if !tally.config.IgnoreWarnings {
			tally.warn("Stopping on warning")
			return ret, err
		}
	}

	return ret, nil
}

func (tally *tally) storeCollectionToFile(coll RSCollection, fileTo string) error {
	var file, err = os.Create(fileTo)
	if err != nil {
		return tally.accessError(fileTo, "Cannot open for writing", err)
	}
	err = coll.StoreTo(file)
	var closeErr = file.Close()
	if err != nil {
		return tally.accessError(fileTo, "Cannot save", err)
	}
	if closeErr != nil {
		return tally.accessError(fileTo, "Cannot close file", closeErr)
	}
	tally.debug("Successfully saved collection to ", fileTo)
	return nil
}

func (tally *tally) loadExistingCollection(fromFile string) (RSCollection, error) {
	var stat, err = os.Stat(fromFile)
	var fileExists bool

	if err != nil {
		tally.debug("Got error from os.Stat(", fromFile, "): ", err)
		if !os.IsNotExist(err) {
			tally.debug("The previous error does not mean file non-existing. Must be some other problem, reporting error")
			return nil, tally.accessError(fromFile, "Cannot stat", err)
		} else {
			tally.debug("The previous error means file does not exist, that is Ok")
			fileExists = false
		}
	} else {
		if !tally.isFile(stat) {
			tally.debug("Path ", fromFile, " is not a regular file, failing")
			return nil, tally.accessError(fromFile, "File is DIRECTORY and cannot be written to", nil)
		}
		fileExists = true
	}

	var coll = NewCollection()
	if fileExists {
		var file *os.File
		file, err = os.Open(fromFile)
		if err != nil {
			return nil, tally.accessError(fromFile, "Cannot open", err)
		}
		err = coll.LoadFrom(file)
		var closeErr = file.Close()
		if err != nil {
			tally.warn("Cannot load file", fromFile)
			if !tally.config.IgnoreWarnings {
				tally.err("Stopping on warning")
				return nil, tally.accessError(fromFile, "Load error", err)
			} else {
				tally.warn("Using empty collection (will rehash files)")
				coll.InitEmpty()
			}

		}
		if closeErr != nil {
			return nil, tally.accessError(fromFile, "Cannot close", err)
		}
		tally.debug("Successfully loaded RSCollection from ", fromFile)
	} else {
		tally.debug("Initializing empty collection since file does not exist")
		coll.InitEmpty()
	}
	return coll, nil
}

func (tally *tally) accessError(filepath string, message string, cause error) error {
	var ret = new(AccessError)

	ret.filepath = filepath
	ret.message = message
	ret.cause = cause

	tally.err(ret)

	return ret
}

func (tally *tally) initLog() {
	tally.loggerDebug = log.New(tally.log, "DEBUG: ", log.Ldate|log.Ltime)
	tally.loggerInfo = log.New(tally.log, "INFO: ", log.Ldate|log.Ltime)
	tally.loggerWarn = log.New(tally.log, "WARN: ", log.Ldate|log.Ltime)
	tally.loggerErr = log.New(tally.log, "ERROR: ", log.Ldate|log.Ltime)
}

func (tally *tally) err(v ...interface{}) {
	if tally.config.LogVerbosity >= 1 {
		tally.loggerErr.Println(v)
	}
}

func (tally *tally) warn(v ...interface{}) {
	if tally.config.LogVerbosity >= 2 {
		tally.loggerWarn.Println(v)
	}
}

func (tally *tally) info(v ...interface{}) {
	if tally.config.LogVerbosity >= 3 {
		tally.loggerInfo.Println(v)
	}
}

func (tally *tally) debug(v ...interface{}) {
	if tally.config.LogVerbosity >= 4 {
		tally.loggerDebug.Println(v)
	}
}

func (tally *tally) isDir(file os.FileInfo) bool {
	return file.Mode().IsDir()
}

func (tally *tally) isFile(file os.FileInfo) bool {
	return file.Mode().IsRegular()
}

func (tally *tally) resolveCollectionFileForDirectory(directory string) (string, error) {
	var context, err = tally.createEvaluationContext(directory)
	if err != nil {
		return "", err
	}
	var ret string
	ret, err = tally.executeTemplate(tally.collectionPathnameTemplate, context)
	if err != nil {
		return "", err
	}

	if ret == "" {
		var tplErr = new(ExpressionError)
		tplErr.expression = tally.config.CollectionPathnameExpression
		tplErr.message = "Evaluates to empty string"
		tally.err(tplErr)
		return "", tplErr
	}

	if ret[0] == '/' {
		tally.debug(ret, "is an absolute path")
	} else {
		ret = filepath.Join(filepath.Dir(directory), ret)
	}

	tally.debug("Resolved collection file for", directory, ":", ret)
	return ret, nil
}

type pathnameEvaluationContext struct {
	path []string // Array of pathname components, for "/1/2/3" it should
                      // be {"1", "2", "3"}
}

func (context *pathnameEvaluationContext) Path(idx int) string {
	var arrIdx = len(context.path)-1 + idx
	if arrIdx<0 || arrIdx>=len(context.path) {
		return ""
	}
	return context.path[arrIdx]
}

func (tally *tally) ensureTemplateCompiled() error {
	if tally.collectionPathnameTemplate == nil {
		var tpl, err = tally.compileTemplate(tally.config.CollectionPathnameExpression)
		if err != nil {
			return err
		}
		tally.collectionPathnameTemplate = tpl
	}
	return nil
}

func (tally *tally) compileTemplate(expr string) (*template.Template, error) {
	var tpl = template.New("template")
	var err error
	tpl, err = tpl.Parse(expr)
	if err != nil {
		var tplErr = new(ExpressionError)
		tplErr.expression = expr
		tplErr.message = "Cannot parse"
		tplErr.cause = err
		tally.err(tplErr)
		return nil, tplErr
	}
	return tpl, nil
}

func (tally *tally) createEvaluationContext(directory string) (TallyPathNameEvalutationContext, error) {
	var ret  = new(pathnameEvaluationContext)
	var normalized, err = filepath.Abs(directory)
	if err != nil {
		return nil, tally.accessError(directory, "Cannot resolve absolute pathname", err)
	}
	ret.path = strings.Split(normalized, string(filepath.Separator))
	return ret, nil
}

func (tally *tally) executeTemplate(tpl *template.Template, context TallyPathNameEvalutationContext) (ret string, err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			tally.err("Error executing template", panicErr)

			var tplErr = new(ExpressionError)
			tplErr.message = "Panic during evaluation"
			tally.err(tplErr)
			ret = ""
			err = tplErr
		}
	}()

	var buf strings.Builder
	err = tpl.Execute(io.Writer(&buf), context)
	if err != nil {
		var tplErr = new(ExpressionError)
		tplErr.message = "Cannot evaluate"
		tplErr.cause = err
		tally.err(tplErr)
		return "", tplErr
	}
	ret = buf.String()
	tally.debug("executeTemplate return", ret)
	return ret, nil
}
