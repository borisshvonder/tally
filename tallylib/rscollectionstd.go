package tallylib

import (
	//	"bufio"
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

func (coll *RSCollectionStd) ByPath(path string) RSCollectionFile {
	return coll.files[path]
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
