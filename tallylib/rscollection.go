package tallylib

import (
	"io"
	"time"
)

type RSCollection interface {
	Update(path string, sha1 string, size uint64,
		timestamp time.Time) RSCollectionFile

	ByPath(path string) RSCollectionFile
	Visit(cb func(RSCollectionFile))

	InitEmpty()
	LoadFrom(in io.Reader) error
	StoreTo(out io.Writer) error
}

type RSCollectionFile interface {
	Path() string
	Sha1() string
	Size() uint64
	Timestamp() time.Time
}
