package tallylib

import (
	"bufio"
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

func (file *RSCollectionFileStd) Path() string {
	return file.path
}

func (file *RSCollectionFileStd) Sha1() string {
	return file.sha1
}

func (file *RSCollectionFileStd) Timestamp() time.Time {
	return file.timestamp
}
