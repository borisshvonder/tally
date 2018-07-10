package main

import (
	"fmt"
	"serv6/tally/tallylib"
	"flag"
	"os"
	"path/filepath"
)

func main() {
	flag.Usage = func() {
		fmt.Println("This program is designed to overcome RetroShare default search limitations. It generates <folder>.rscollection file for each folder encountered, forming so-called 'collection tree', that is a tree of .rscollection files referencing each other. These <folder>.rscollection files serve as RetroShare 'folders' that can be found using RetroShare search without revealing folder structure to any peers directly.")
		var me = os.Args[0]
		fmt.Fprintf(flag.CommandLine.Output(), "USAGE: %s [options] [folder1, folder2, ...]\n", me)
		flag.PrintDefaults()
		fmt.Printf(`EXAMPLES
	tallycli -IgnoreWarnings /my/audiobooks
		Generate collection tree for all subfolders of /my/audiobooks including /my/audiobooks.rscollection
	tallycli -UpdateParents /my/audiobooks/Heller/Catch-22
		Update /my/audiobooks/Heller/Catch-22.rscollection, /my/audiobooks/Heller.rscollection and /my/audiobooks.rscollection
	tallycli -LogVerbosity 0 -UpdateRecursive=false /my/audiobooks/Heller/Something-Happened
		Quietly update just /my/audiobooks/Heller/Something-Happened.rscollection
`)
	}

	var tally = tallylib.NewTally()
	var config = tally.GetConfig()

	flag.BoolVar(&config.IgnoreWarnings, "IgnoreWarnings", false, "ignore warnings, should be fine for most usecases")
	flag.BoolVar(&config.RemoveExtraFiles, "RemoveExtraFiles", false, "remove any files referenced in .rscollections that tool does not handle")
	flag.BoolVar(&config.UpdateParents, "UpdateParents", false, "when updating a directory, also update parent directories")
	flag.BoolVar(&config.ForceUpdate, "ForceUpdate", false, "rehash files regardless of their sizes and timestamps")
	flag.IntVar(&config.LogVerbosity, "LogVerbosity", 3, "log level [0-4], default:3")

	var UpdateRecursive bool
	flag.BoolVar(&UpdateRecursive, "UpdateRecursive", true, "update folders recursively")
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
