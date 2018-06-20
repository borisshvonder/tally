package tallylib

import (
	"bufio"
	"time"
)

type RScollection interface {
	Update(path string, sha1 string, timestamp time.Time) RSCollectionFile
	ByPath(path string) RSCollectionFile
	Visit(cb func(RSCollectionFile))

	InitEmpty()
	LoadFrom(in bufio.Reader)
	StoreTo(out bufio.Writer)
}

type RSCollectionFile interface {
	Path() string
	Sha1() string
	Timestamp() time.Time
}
