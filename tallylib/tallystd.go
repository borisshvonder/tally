package tallylib

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Holds settings
type tally struct {
	config      TallyConfig
	log         io.Writer
	loggerDebug *log.Logger
	loggerInfo  *log.Logger
	loggerErr   *log.Logger
	loggerWarn  *log.Logger
}

func NewTally() Tally {
	var ret = new(tally)
	ret.config.logVerbosity = 3
	ret.SetLog(ioutil.Discard)
	return ret
}

func (tally *tally) GetConfig() TallyConfig {
	return tally.config
}

func (tally *tally) SetConfig(cfg TallyConfig) {
	tally.config = cfg
}

func (tally *tally) SetLog(logfile io.Writer) {
	tally.log = logfile
	tally.initLog()
}

func (tally *tally) UpdateRecursive(directory string) (bool, error) {
	tally.info("UpdateRecursive(", directory, ")")
	var err = tally.assertDirectory(directory)
	if err != nil {
		return false, err
	}

	var files []os.FileInfo
	files, err = tally.listDirectory(directory)

	var ret = false	
	var changed = false

	tally.debug("Stage1: updating subdirectories")
	
	for _, file := range files {
		if file.IsDir() {
			tally.debug("Invoking UpdateRecursive(", file.Name(), ")")
			var fullpath = filepath.Join(directory, file.Name())
			changed, err = tally.UpdateRecursive(fullpath)
			ret = ret || changed
			if err != nil {
				return ret, err
			}
		} else {
			tally.debug("Skipping file", file.Name())
		}
	}


	tally.debug("Stage2: update", directory)

	changed, err = tally.UpdateSingleDirectory(directory)
	ret = ret || changed
	if err != nil {
		return ret, err
	}

	if tally.config.updateParents {
		tally.debug("Stage3: updating parents")
		changed, err = tally.updateParents(directory)
		ret = ret || changed
	}

	return ret, err
}

func (tally *tally) updateParents(directory string) (bool, error) {
	var ret = false
	var err error
	
	for parent := filepath.Dir(directory); err == nil && parent != ""; parent = filepath.Dir(parent) {
		var collectionFile = resolveCollectionFileForDirectory(parent)
		var stat, err = os.Stat(collectionFile)
		if err == nil {
			if stat.IsDir() {
				tally.info("Stopping updating patents, file", collectionFile, "is a directory")
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
	tally.info("UpdateSingleDirectory(", directory, ")")
	var err = tally.assertDirectory(directory)
	if err != nil {
		return false, err
	}

	// In order to remove files which no longer exists from the collection,
	// we use 2 collections here: oldColl and newColl. This is actually
	// does not consume too much memory since collections store interfaces
	// (pointers) to real data. Regardless, the data is very small anyway
	var collectionFile = resolveCollectionFileForDirectory(directory)
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
	files, err = tally.listDirectory(directory)
	var ret = false

	for _, file := range files {
		if file.IsDir() {
			tally.debug("Skipping", file.Name(), "because it is a directory")
		} else {
			var changed bool
			var name = file.Name()
			tally.debug("Working on file", name)

			var oldFile = oldColl.ByName(name)
			oldColl.RemoveFile(name)

			if oldFile != nil {
				newColl.UpdateFile(oldFile)
			}
			changed, err = tally.updateFile(directory, name, newColl)
			if err != nil {
				return ret, err
			}
			if changed {
				ret = true
				tally.info("Detected change in", name)
			}
		}
	}

	// files left in old collection could mean files were removed from disk
	if tally.config.removeExtraFiles {
		ret = ret || oldColl.Size() > 0
	} else {
		var changed = tally.removeMissingFiles(directory, oldColl, newColl)
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
	if !stat.IsDir() {
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
	var ret, err = updateFile(coll, filename, fullpath, tally.config.forceUpdate)

	if err != nil {
		// Failure to update single file is not critical
		tally.warn("Could not update", fullpath, err)
		if !tally.config.ignoreWarnings {
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
		if stat.IsDir() {
			tally.debug("Path ", fromFile, " is a directory, failing")
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
			if !tally.config.ignoreWarnings {
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
	if tally.config.logVerbosity >= 1 {
		tally.loggerErr.Println(v)
	}
}

func (tally *tally) warn(v ...interface{}) {
	if tally.config.logVerbosity >= 2 {
		tally.loggerWarn.Println(v)
	}
}

func (tally *tally) info(v ...interface{}) {
	if tally.config.logVerbosity >= 3 {
		tally.loggerInfo.Println(v)
	}
}

func (tally *tally) debug(v ...interface{}) {
	if tally.config.logVerbosity >= 4 {
		tally.loggerDebug.Println(v)
	}
}

func resolveCollectionFileForDirectory(directory string) string {
	dir, file := filepath.Split(filepath.Clean(directory))
	if file == "." {
		file = ""
	}
	return filepath.Join(dir, file+".rscollection")
}
