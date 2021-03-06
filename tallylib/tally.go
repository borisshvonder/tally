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
	// if addChildren=true, then also add all subdirectories and files
	// to this directory collection
	UpdateSingleDirectory(directory string, addChildren bool) (bool, error)

	// Update all subdirectories and all parent directories (if any)
	// returns true if made any changes or false if no changes done
	// minDig specifies at which depth start to generate rscollection 
        // files (0 is default, generate from top)
	// maxDig specifies at wich depth stop generating rscollection files
	// (default is -1, means generate until bottom)
	// When maxDig is reached, don't generate additional rscollections,
	// just put all subdirectories in the rscollection file at this depth
	//
	// Some corner cases:
	// * UpdateRecursive(directory, 0, 0) is same as 
        //   UpdateSingleDirectory(directory, true)
        // * UpdateRecursive(directory, 0, -1) always recurse to bottom
	UpdateRecursive(directory string, minDig, maxDig int) (bool, error)

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

	// This expression is evaluated using same framework ad 
	// CollectionPathnameExpression
	// By default it is empty, which means all files are simply put into
	// collections at the top of the collection.
	// If it is not empty, then files are put under the root path
	// specified, for example, if CollectionRootPathExpression 
	// evaluates to "Music/MP3", then file "Sepultura/1993/01.mp3"
	// will be put into collection as "Music/MP3/Sepultura/1993/01.mp3".
	// Note: the path separator is '/' regardless of OS.
	CollectionRootPathExpression string
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
	fullpath string // path to file
	message  string // additional message
	cause    error  // underlying error, if any
}

func (e *AccessError) Error() string {
	return e.fullpath + " " + e.message + " " + e.cause.Error()
}

type ExpressionError struct {
	expression string // Expression caused error
	message    string // Error message
	cause      error  // underlying error, if any
}

func (e *ExpressionError) Error() string {
	return e.expression + ": " + e.message + " " + e.cause.Error()
}
