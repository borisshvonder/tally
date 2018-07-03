package tallylib

import (
	"io"
)

type TallyConfig struct {
	forceUpdate bool

	// By default, tally stops on errors but continues on warnings.
	// Set this to true and tally will act like coward.
	stopOnWarnings bool

	// Log verbosity.
	//   0 means do not log anything,
	//   1 - only errors,
	//   2 - warnings + errors
	//   3 - warnings + errors + info (default)
	//   4 - warnings + errors + info + debug
	// more - increase logging more and more
	logVerbosity int
}

type Tally interface {
	GetConfig() TallyConfig
	SetConfig(cfg TallyConfig)

	// Where to log stuff, by default don't write anywhere
	SetLog(log io.Writer)

	UpdateSingleDirectory(directory string) error

	// Update all subdirectories and all parent directories (if any)
	UpdateRecursive(directory string) error
}

type CollectionFileAccessError struct {
	filepath string // path to file
	message  string // additional message
	cause    error  // underlying error, if any
}

func (e *CollectionFileAccessError) Error() string {
	return e.filepath + " " + e.message + " " + e.cause.Error()
}
