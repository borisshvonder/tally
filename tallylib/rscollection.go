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

	// Store this collection to .rscollection XML file.
	StoreTo(out io.Writer) error

	// Update record in collection
	// Note that name does not have to be an actual file path.
	// It is typically a file path, relative to current directory
	Update(name, sha1 string, size int64, timestamp time.Time) RSCollectionFile

	// Update collection with exising RSCollectionFile, could be used
	// as an effective 'copy' operation
	UpdateFile(file RSCollectionFile)

	// Remove file from collection
	RemoveFile(name string)

	// Find exising record by file name (not full path)
	// Note that name does not have to be an actual file path.
	// It is typically a file path, relative to current directory
	ByName(name string) RSCollectionFile

	// Invode callback for every file stored in this collection.
	// Order is not guaranteed
	Visit(cb func(RSCollectionFile))

	// Returns collection size
	Size() int
}

type RSCollectionFile interface {
	Name() string         // a RELATIVE path or name
	Sha1() string         // sha1 encoded as lowercase hex letters
	Size() int64          // size of file (0 if unknown)
	Timestamp() time.Time // file mod time
}
