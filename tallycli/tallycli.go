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

COLLECTION ROOT PATH EXPRESSIONS
	Default behavior is simply put files found into collection at the
	top, for path
	/docs/file.txt
	the /docs.rscollection will contain file "file.txt'

	However, sometimes it is desirable to put all files inside a collection
	under common root. For example:

	$ tally -CollectionRootPathExpression="MyDocs/Texts" /docs
	will put file "MyDocs/Texts/file.txt" into the collection.

	The most common example of this is when sharing music:


	/music/Artist1/Album1
	/music/Artist2/Album1

	$ tally /music 

	will generate 
          /music/Artist1.rscollection and 
          /music/Artist2.rscollectionwhich 
	which both will contain "Album1.rscollection" files. 
	When user downloads both collections, files will chash between each
	other.

	To prevent this, we could do:

	$ tally -CollectionRootPathExpression="{{.Path -1}}" /music
	which will generate collections
	  /music/Artist1.rscollection containing
            "Artist1/Album1.rscollection"
	  and 
	  /music/Artist2.rscollection containing
	    "Artist2/Album2.rscollection

DIG DEPTH
	By default, tally creates single .rscollection for each directory in
	a tree. Sometimes it might not be desirable. For example, given this
	typical music tree:

	/music/Sepultura/1993_Roots/Lyrics

	it is not quite makes sense creating collections 
	/music/Sepultura/1993_Roots.rscollection 
	and
	/music/Sepultura/1993_Roots/Lyrics.rscollection 

	For that reason, the default behavior could be changed using 
	-MinDig and -MaxDig parameters. It is best explained by an example:

	tally -MinDig=1 -MaxDig=2 /music
	will create collections:
	/music/Sepultura.rscollection
	/music/Sepultura/1993_Roots.rscollection (this one will also include 
           Lyrics folder.)
        It won't, however, create collections:
        /music.rscollecton (because depth=0 < MinDig)
        /music/Sepultura/1993_Roots/Lyrics.rscollection (depth=3 > MaxDig)

	To make names better, we can do
	 tally -MinDig=1 -MaxDig=2 \
	  -CollectionPathnameExpression="{{.Path -1}}-{{.Path 0}}.rscollection"\
	  /music

	which will create collections:
	/music/music-Sepultura.rscollection
	/musci/Sepultura/Sepultura-1993_Roots.rscollection

	The corner case is -MaxDig=0, in this case tally will generate just
	one large collection containing all files in a given folder.
	

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
	var MinDig, MaxDig int

	flag.BoolVar(&config.IgnoreWarnings, "IgnoreWarnings", false, "ignore warnings, should be fine for most usecases")
	flag.BoolVar(&config.RemoveExtraFiles, "RemoveExtraFiles", false, "remove any files referenced in .rscollections that tool does not handle")
	flag.BoolVar(&config.UpdateParents, "UpdateParents", true, "when updating a directory, also update parent directories")
	flag.BoolVar(&config.ForceUpdate, "ForceUpdate", false, "rehash files regardless of their sizes and timestamps")
	flag.IntVar(&config.LogVerbosity, "LogVerbosity", 3, "log level [0-4], default:3")

	var UpdateRecursive bool
	flag.BoolVar(&UpdateRecursive, "UpdateRecursive", true, "update folders recursively")

	flag.StringVar(&config.CollectionPathnameExpression, "CollectionPathnameExpression", "{{.Path 0}}.rscollection", 	
		"Template expression for resolving .rscollection file name, see COLLECTION EXPRESSIONS")

	flag.StringVar(&config.CollectionRootPathExpression, "CollectionRootPathExpression", "", 	
		"Template expression for resolving root path in rscollection, see COLLECTION ROOT PATH EXPRESSIONS")

	flag.IntVar(&MinDig, "MinDig", 0, "create intermediate .rscollections only from this folder level down. See DIG DEPTH")
	flag.IntVar(&MaxDig, "MaxDig", -1, "create intermediate .rscollections up to this depth, -1 means infinite. See DIG DEPTH")

	flag.Parse()

	tally.SetConfig(config)
	tally.SetLog(os.Stdout)

	for _, path := range flag.Args() {
		path = filepath.Clean(path)
		var err error
		if UpdateRecursive {
			_, err = tally.UpdateRecursive(path, MinDig, MaxDig)
		} else {
			_, err = tally.UpdateSingleDirectory(path, false)
		}
		if err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
	}
}
