package main

import (
	"fmt"
	"serv6/tally/tallylib"
	"flag"
	"os"
	"path/filepath"
)

func main() {
	var tally = tallylib.NewTally()
	var config = tally.GetConfig()

	flag.BoolVar(&config.IgnoreWarnings, "IgnoreWarnings", false, "ignore warnings, should be fine for most usecases")
	flag.BoolVar(&config.RemoveExtraFiles, "RemoveExtraFiles", false, "remove any files referenced in .rscollections that tool does not handle")
	flag.BoolVar(&config.UpdateParents, "UpdateParents", false, "when updating a directory, also update parent directories")
	flag.BoolVar(&config.ForceUpdate, "ForceUpdate", false, "rehash files regardless of their sizes and timestamps")
	flag.IntVar(&config.LogVerbosity, "LogVerbosity", 3, "log level [0-4], default:3")

	var UpdateRecursive bool
	flag.BoolVar(&UpdateRecursive, "UpdateRecursive", true, "update folders recursively (default:true)")
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
