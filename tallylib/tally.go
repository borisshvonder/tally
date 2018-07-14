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
//  IgnoreWarnings=true for most of the usecases
// Consider also setting RemoveExtraFiles=true and UpdateParents=true
type TallyConfig struct {
	// By default, tally stops on both errors and warnings. 
	// Typically, ignoring warnings is pretty safe, but you may get 
	// your .recollection files overwritten.
	IgnoreWarnings bool

	// Remove any existing entries in .rscollection files that
	// refer existing files in deeper in subtree, ex
	// name="mydir/myotherdir/somegile"
	// By default, tally keeps them untouched removing only ones
	// which not present
	RemoveExtraFiles bool

	// Useful when you have an exising tree of .rscollection files but
	// want to update just specific subfolder. When UpdateParents=true,
	// tally will go up your subfolders and update any parent folders 
	// with .rscollection files present
	UpdateParents bool

	// Force hash recalculation even if it looks like files were not
	// updated
	ForceUpdate bool

	// Log verbosity.
	//   0 means do not log anything,
	//   1 - only errors,
	//   2 - warnings + errors
	//   3 - warnings + errors + info (default)
	//   4 - warnings + errors + info + debug
	// more - increase logging more and more
	LogVerbosity int

	// This expression is evalated by standard go text/template framework
	// when resolving collection file name from directory(for which 
	// collection is mad) path.
	// By default it is simply "{{.Path(0)}}.rscollection", which makes
	// collection file name same as directory name. You can specify
	// something more fancy, like for ex 
	// "{{.Path(-1)}}-{{.Path(0)}}.rscollection", refer to 
	// TallyPathNameEvalutationContext for available template functions
	CollectionPathnameExpression string

}

// When resolving collection name (see TallyConfig.CollectionPathnameExpression)
// the template expression is executed against this interface.
type TallyPathNameEvalutationContext interface {
	// Returns path component of the directory
	// idx=0 means directory name,
	// idx=-1 means parent directory name and so on, for example, for
	// directory="/my/directory/1/2":
	//    Path(0)  returns "2"
	//    Path(-1) returns "1"
	//    Path(-2) returns "directory"
	//    Path(-3) returns "my"
	//    Path(-4) returns ""
	// For any non-existing component empty string will be returned: 
	//    Path(-4) returns ""
	//    Path(1)  returns ""
	Path(idx int) string
}

type AccessError struct {
	filepath string // path to file
	message  string // additional message
	cause    error  // underlying error, if any
}

func (e *AccessError) Error() string {
	return e.filepath + " " + e.message + " " + e.cause.Error()
}

type ExpressionError struct {
	expression string // Expression caused error
	message    string // Error message
	cause      error  // underlying error, if any
}

func (e *ExpressionError) Error() string {
	return e.expression + ": " + e.message + " " + e.cause.Error()
}
