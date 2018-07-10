# tally
Retroshare indexing tool

This program is designed to overcome RetroShare default search limitations. It generates <folder>.rscollection file for each folder encountered, forming so-called 'collection tree', that is a tree of .rscollection files referencing each other. These <folder>.rscollection files serve as RetroShare 'folders' that can be found using RetroShare search without revealing folder structure to any peers directly.

to download and build: 

$ go get -v github.com/borisshvonder/tally

$ cd $GOPATH/src/github.com/borisshvonder/tally/tallycli && go build

You need to set up you go language environment correctly.

Or, If you are very lazy and trust random binaries, go to [release](release/) folder
