package tallylib

import (
	"io"
	"time"
)

type RSCollectionStd struct {
	files map[string]RSCollectionFile
}

type RSCollectionFileStd struct {
	path      string
	sha1      string
	timestamp time.Time
}

func (coll *RSCollectionStd) InitEmpty() {
	coll.files = make(map[string]RSCollectionFile)
}

func (coll *RSCollectionStd) StoreTo(out io.Writer) {
}

func (coll *RSCollectionStd) Update(
	path string,
	sha1 string,
	timestamp time.Time) RSCollectionFile {

	var file = new(RSCollectionFileStd)
	file.path = path
	file.sha1 = sha1
	file.timestamp = timestamp

	coll.files[path] = file

	return file
}

func (coll *RSCollectionStd) Visit(cb func(RSCollectionFile)) {
	for _, v := range coll.files {
		cb(v)
	}
}

func (coll *RSCollectionStd) ByPath(path string) RSCollectionFile {
	return coll.files[path]
}

func (coll *RSCollectionStd) LoadFrom(in io.Reader) {
	coll.files = make(map[string]RSCollectionFile)
}

func (file *RSCollectionFileStd) Path() string {
	return file.path
}

func (file *RSCollectionFileStd) Sha1() string {
	return file.sha1
}

func (file *RSCollectionFileStd) Timestamp() time.Time {
	return file.timestamp
}
