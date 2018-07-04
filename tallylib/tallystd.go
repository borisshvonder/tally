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

func (tally *tally) UpdateSingleDirectory(directory string) (bool, error) {
	var collectionFile = directory + ".rscollection"

	var coll, err = tally.loadExistingCollection(collectionFile)
	if err != nil {
		return false, err
	}

	var files []os.FileInfo
	files, err = ioutil.ReadDir(directory)
	var ret = false

	for _, file := range files {
		if !file.IsDir() {
			var changed bool
			changed, err = tally.updateFile(directory, file.Name(), coll)
			ret = changed || ret
			if err != nil {
				return ret, err
			}
		}
	}

	return ret, nil
}

func (tally *tally) updateFile(directory, filename string, coll RSCollection) (bool, error) {
	var fullpath = filepath.Join(directory, filename)
	var ret, err = updateFile(coll, fullpath, filename, tally.config.forceUpdate)

	if err != nil {
		// Failure to update single file is not critical
		tally.warn("Could not update ", fullpath, err)
		if tally.config.stopOnWarnings {
			tally.warn("Stopping on warning")
			return ret, err
		}
	}

	return ret, nil
}

func (tally *tally) loadExistingCollection(fromFile string) (RSCollection, error) {
	var stat, err = os.Stat(fromFile)

	if err != nil && os.IsNotExist(err) {
		return nil, tally.collectionFileError(fromFile, "Cannot stat", err)
	} else if err == nil {
		if stat.IsDir() {
			return nil, tally.collectionFileError(fromFile, "File is DIRECTORY and cannot be written to", nil)
		}
	}

	var file *os.File
	file, err = os.Open(fromFile)
	if err != nil {
		return nil, tally.collectionFileError(fromFile, "Cannot open", err)
	}
	var coll = NewCollection()
	err = coll.LoadFrom(file)
	var closeErr = file.Close()
	if err != nil {
		tally.warn("Cannot load file", fromFile)
		if tally.config.stopOnWarnings {
			tally.err("Stopping on warning")
			return nil, tally.collectionFileError(fromFile, "Load error", err)
		} else {
			tally.warn("Using empty collection (will rehash files)")
			coll.InitEmpty()
		}

	}
	if closeErr != nil {
		return nil, tally.collectionFileError(fromFile, "Cannot close", err)
	}
	return coll, nil
}

func (tally *tally) collectionFileError(filepath string, message string,
	cause error) error {

	var ret = new(CollectionFileAccessError)

	ret.filepath = filepath
	ret.message = message
	ret.cause = cause

	tally.err(ret)

	return ret
}

func (tally *tally) UpdateRecursive(directory string) (bool, error) {
	panic("NIY")
}

func (tally *tally) initLog() {
	tally.loggerDebug = log.New(tally.log, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	tally.loggerInfo = log.New(tally.log, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	tally.loggerWarn = log.New(tally.log, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	tally.loggerErr = log.New(tally.log, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
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
		tally.loggerErr.Println(v)
	}
}
