package tallylib

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"time"
)

func New() RSCollection {
	var ret = new(collection)
	return ret
}

type collection struct {
	files map[string]RSCollectionFile
}

type file struct {
	path      string
	sha1      string
	timestamp time.Time
}

func (coll *collection) InitEmpty() {
	coll.files = make(map[string]RSCollectionFile)
}

func (coll *collection) StoreTo(out io.Writer) {
}

func (coll *collection) Update(
	path string,
	sha1 string,
	timestamp time.Time) RSCollectionFile {

	var file = new(file)
	file.path = path
	file.sha1 = sha1
	file.timestamp = timestamp

	coll.files[path] = file

	return file
}

func (coll *collection) Visit(cb func(RSCollectionFile)) {
	for _, v := range coll.files {
		cb(v)
	}
}

func (coll *collection) ByPath(path string) RSCollectionFile {
	return coll.files[path]
}

type XmlRsCollection struct {
	XMLName xml.Name   `xml:"RsCollection"`
	Files   []*XmlFile `xml:"File"`
}

type XmlFile struct {
	XMLName xml.Name `xml:"File"`
	Sha1    string   `xml:"sha1,attr"`
	Name    string   `xml:"name,attr"`
	Size    uint64   `xml:"size,attr"`
	Updated string   `xml:"updated,attr"`
}

func (coll *collection) LoadFrom(in io.Reader) {
	var data, _ = ioutil.ReadAll(in)
	var parsed = new(XmlRsCollection)

	var _ = xml.Unmarshal(data, parsed)

	coll.files = make(map[string]RSCollectionFile)

	for i := range parsed.Files {
		var xmlFile = parsed.Files[i]
		var timestamp, _ = time.Parse(time.RFC3339, xmlFile.Updated)
		coll.Update(xmlFile.Name, xmlFile.Sha1, timestamp)
	}
}

func (file *file) Path() string {
	return file.path
}

func (file *file) Sha1() string {
	return file.sha1
}

func (file *file) Timestamp() time.Time {
	return file.timestamp
}
