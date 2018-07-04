package tallylib

import (
	"io"
	"time"
)

// Represents .rscollection file
type RSCollection interface {
	// Initialize empty collection. You MUST call either this or LoadFrom
	// before using the object
	InitEmpty()

	// Load collection from .rscollection XML file.
	LoadFrom(in io.Reader) error
	StoreTo(out io.Writer) error

	// Update record in collection
	// Note that name does not have to be an actual file path.
	// It is typically a file path, relative to current directory
	Update(name, sha1 string, size int64, timestamp time.Time) RSCollectionFile

	// Find exising record by file name (not full path)
	// Note that name does not have to be an actual file path.
	// It is typically a file path, relative to current directory
	ByName(name string) RSCollectionFile

	// Invode callback for every file stored in this collection.
	// Order is not guaranteed
	Visit(cb func(RSCollectionFile))
}

type RSCollectionFile interface {
	Name() string
	Sha1() string
	Size() int64
	Timestamp() time.Time
}
