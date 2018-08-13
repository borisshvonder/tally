package main

import (
	"fmt"
	"github.com/borisshvonder/tally/tallylib"
	"flag"
	"os"
	"path/filepath"
	"runtime"
)

var version string // set by linker

func main() {
	flag.Usage = func() {
		fmt.Println("This program is designed to overcome RetroShare default search limitations. It generates <folder>.rscollection file for each folder encountered, forming so-called 'collection tree', that is a tree of .rscollection files referencing each other. These <folder>.rscollection files serve as RetroShare 'folders' that can be found using RetroShare search without revealing folder structure to any peers directly.")
		var me = os.Args[0]
		fmt.Fprintf(flag.CommandLine.Output(), "USAGE: %s [options] [folder1, folder2, ...]\n", me)
		flag.PrintDefaults()
		fmt.Printf(`EXAMPLES
	tally -IgnoreWarnings /my/audiobooks
		Generate collection tree for all subfolders of /my/audiobooks 
		including /my/audiobooks.rscollection
	tally -UpdateParents /my/audiobooks/Heller/Catch-22
		Update /my/audiobooks/Heller/Catch-22.rscollection, 
		/my/audiobooks/Heller.rscollection and /my/audiobooks.rscollection

	tally -LogVerbosity 0 -UpdateRecursive=false \
			/my/audiobooks/Heller/Something-Happened

		Quietly update just 
		/my/audiobooks/Heller/Something-Happened.rscollection

COLLECTION EXPRESSIONS

	By default, tally assigns collection file names same as respective
	directory names. For example, the /path/to/dir directory collection
	will be resolved as /path/to/dir.rscollection.

	However, some users find it useful to customize this process with
	-CollectionPathnameExpression flag. Default setting for this flag
	precisely follows default behavior: "{{.Path 0}}.rscollection" resolves 
	to last path component (.Path 0) in directory pathname, ex "dir" 
	for "/path/to/dir" plus ".rscollection" string

	Currently templating only allows you selecting parent directories and
	it is easier to explain usng following example:

	* assuming we have a directory /music/Depeche Mode/Violator_1993 ...

	* -CollectionPathnameExpression="{{.Path -1}}-{{.Path 0}}.rscollection"
          will produce value "Depeche Mode-Violator_1993.rscollection"

	* ...Expression="{{.Path -2}}-{{.Path -1}}-{{.Path 0}}.rscollection"
	  will produce value "Music-Depeche Mode-Violator_1993.rscollection"

	The expression can contain slashes and you can use them to store all
	your .rscollections in a separate folder. However, this is discouraged
	since then parent collections will not include child collection files:

	* -CollectionPathnameExpression="/collections/{{.Path 0}}.rscollection"

BUGS
	In order to efficiently detect if file needs it's sha1 recalculated, 
	this tool stores file modification time in .rscollection file as 
	non-standard (unsupported by RetroShare) attribute "updated". 
	So far, RetroShare does not seem to care, but, in future, it may stop
	handling such files.	
`)
		fmt.Printf("OS: %s\nArchitecture: %s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("Version:", version)
	}

	var tally = tallylib.NewTally()
	var config = tally.GetConfig()

	flag.BoolVar(&config.IgnoreWarnings, "IgnoreWarnings", false, "ignore warnings, should be fine for most usecases")
	flag.BoolVar(&config.RemoveExtraFiles, "RemoveExtraFiles", false, "remove any files referenced in .rscollections that tool does not handle")
	flag.BoolVar(&config.UpdateParents, "UpdateParents", true, "when updating a directory, also update parent directories")
	flag.BoolVar(&config.ForceUpdate, "ForceUpdate", false, "rehash files regardless of their sizes and timestamps")
	flag.IntVar(&config.LogVerbosity, "LogVerbosity", 3, "log level [0-4], default:3")

	var UpdateRecursive bool
	flag.BoolVar(&UpdateRecursive, "UpdateRecursive", true, "update folders recursively")

	flag.StringVar(&config.CollectionPathnameExpression, "CollectionPathnameExpression", "{{.Path 0}}.rscollection", 	
		"Template expression for resolving .rscollection file name, see COLLECTION EXPRESSIONS")

	flag.Parse()

	tally.SetConfig(config)
	tally.SetLog(os.Stdout)

	for _, path := range flag.Args() {
		path = filepath.Clean(path)
		var err error
		if UpdateRecursive {
			_, err = tally.UpdateRecursive(path)
		} else {
			_, err = tally.UpdateSingleDirectory(path)
		}
		if err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
	}
}
