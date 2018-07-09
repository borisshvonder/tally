package tallylib

import (
	"io"
)

// Facade interface for the library
type Tally interface {
	// Get current configuration (default config if haven't ever called
	// SetConfig)
	GetConfig() TallyConfig

	// Set new configuration
	SetConfig(cfg TallyConfig)

	// returns true if made any changes or false if no changes done
	UpdateSingleDirectory(directory string) (bool, error)

	// Update all subdirectories and all parent directories (if any)
	// returns true if made any changes or false if no changes done
	UpdateRecursive(directory string) (bool, error)

	// Where to log stuff, by default don't write anywhere
	SetLog(log io.Writer)
}

// Tally configuration
// Default options are overly safe, you probably want to set
//  ignoreWarnings=true for most of the usecases
// Consider also setting removeExtraFiles=true and updateParents=true
type TallyConfig struct {
	// By default, tally stops on both errors and warnings. 
	// Typically, ignoring warnings is pretty safe, but you may get 
	// your .recollection files overwritten.
	ignoreWarnings bool

	// Remove any existing entries in .rscollection files that
	// refer existing files in deeper in subtree, ex
	// name="mydir/myotherdir/somegile"
	// By default, tally keeps them untouched removing only ones
	// which not present
	removeExtraFiles bool

	// Useful when you have an exising tree of .rscollection files but
	// want to update just specific subfolder. When updateParents=true,
	// tally will go up your subfolders and update any parent folders 
	// with .rscollection files present
	updateParents bool

	// Force hash recalculation even if it looks like files were not
	// updated
	forceUpdate bool

	// Log verbosity.
	//   0 means do not log anything,
	//   1 - only errors,
	//   2 - warnings + errors
	//   3 - warnings + errors + info (default)
	//   4 - warnings + errors + info + debug
	// more - increase logging more and more
	logVerbosity int
}

type AccessError struct {
	filepath string // path to file
	message  string // additional message
	cause    error  // underlying error, if any
}

func (e *AccessError) Error() string {
	return e.filepath + " " + e.message + " " + e.cause.Error()
}

