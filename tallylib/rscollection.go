package tallylib

import (
	"io"
	"time"
)

type RSCollection interface {
	Update(path string, sha1 string, timestamp time.Time) RSCollectionFile
	ByPath(path string) RSCollectionFile
	Visit(cb func(RSCollectionFile))

	InitEmpty()
	LoadFrom(in io.Reader)
	StoreTo(out io.Writer)
}

type RSCollectionFile interface {
	Path() string
	Sha1() string
	Timestamp() time.Time
}
