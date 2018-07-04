package tallylib

import (
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"time"
)

const xmlHeader = "<!DOCTYPE RsCollection>\n"

func NewCollection() RSCollection {
	var ret = new(collection)
	return ret
}

type collection struct {
	files map[string]RSCollectionFile
}

type file struct {
	path      string
	sha1      string
	size      int64
	timestamp time.Time
}

func (coll *collection) InitEmpty() {
	coll.files = make(map[string]RSCollectionFile)
}

func (coll *collection) Update(path, sha1 string, size int64, timestamp time.Time) RSCollectionFile {
	var file = new(file)
	file.path = path
	file.sha1 = sha1
	file.size = size
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
	Size    int64    `xml:"size,attr"`
	Updated string   `xml:"updated,attr"`
}

func (coll *collection) LoadFrom(in io.Reader) error {
	var data, _ = ioutil.ReadAll(in)
	var parsed = new(XmlRsCollection)

	var err = xml.Unmarshal(data, parsed)
	if err != nil {
		return err
	}

	var errs string

	coll.files = make(map[string]RSCollectionFile)

	for i := range parsed.Files {
		var xmlFile = parsed.Files[i]
		var file, err = xmlFileToStd(xmlFile)
		if err != nil {
			errs += err.Error() + " "
		}
		coll.files[file.path] = file
	}

	if errs != "" {
		err = errors.New(errs)
	}

	return err
}

func (coll *collection) StoreTo(out io.Writer) error {
	var xmlColl = new(XmlRsCollection)
	xmlColl.Files = make([]*XmlFile, len(coll.files))

	var i = 0
	for _, v := range coll.files {
		xmlColl.Files[i] = stdFileToXml(v)
		i += 1
	}

	var _, err = out.Write([]byte(xmlHeader))

	var data []byte
	data, err = xml.Marshal(xmlColl)

	if err == nil {
		_, err = out.Write(data)
	}

	return err
}

func xmlFileToStd(xmlFile *XmlFile) (*file, error) {
	var ret = new(file)
	ret.path = xmlFile.Name
	ret.sha1 = xmlFile.Sha1
	ret.size = xmlFile.Size
	var err error

	if xmlFile.Updated != "" {
		ret.timestamp, err = time.Parse(time.RFC3339Nano,
			xmlFile.Updated)
	}

	return ret, err
}

func stdFileToXml(file RSCollectionFile) *XmlFile {
	var ret = new(XmlFile)
	ret.Name = file.Path()
	ret.Sha1 = file.Sha1()
	ret.Size = file.Size()
	var timestamp = file.Timestamp()
	if timestamp != (time.Time{}) {
		ret.Updated = timestamp.Format(time.RFC3339Nano)
	}
	return ret
}

func (file *file) Path() string {
	return file.path
}

func (file *file) Sha1() string {
	return file.sha1
}

func (file *file) Size() int64 {
	return file.size
}

func (file *file) Timestamp() time.Time {
	return file.timestamp
}
