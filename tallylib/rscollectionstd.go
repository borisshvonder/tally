package tallylib

import (
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"time"
	"strings"
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
	name      string
	sha1      string
	size      int64
	timestamp time.Time
}

func (coll *collection) InitEmpty() {
	coll.files = make(map[string]RSCollectionFile)
}

func (coll *collection) Size() int {
	return len(coll.files)
}

func (coll *collection) Update(name, sha1 string, size int64, timestamp time.Time) RSCollectionFile {
	var file = new(file)
	file.name = name
	file.sha1 = sha1
	file.size = size
	file.timestamp = timestamp

	coll.files[name] = file

	return file
}

func (coll *collection) RemoveFile(name string) {
	delete(coll.files, name)
}

func (coll *collection) UpdateFile(file RSCollectionFile) {
	coll.files[file.Name()] = file
}

func (coll *collection) Visit(cb func(RSCollectionFile)) {
	for _, v := range coll.files {
		cb(v)
	}
}

func (coll *collection) ByName(name string) RSCollectionFile {
	return coll.files[name]
}

type XmlRsCollection struct {
	XMLName xml.Name   `xml:"RsCollection"`
	Directories []*XmlDirectory `xml:"Directory"`
	Files   []*XmlFile `xml:"File"`
}

type XmlDirectory struct {
	XMLName xml.Name `xml:"Directory"`
	Name    string   `xml:"name,attr"`
	Directories []*XmlDirectory `xml:"Directory"`
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

	var errs strings.Builder
	coll.files = make(map[string]RSCollectionFile)

	coll.loadDirectories("", parsed.Directories, &errs)
	coll.loadFiles("", parsed.Files, &errs)

	var errsStr = errs.String()
	if errsStr != "" {
		return errors.New(errsStr)
	} else {
		return nil
	}
}

func (coll *collection) loadDirectories(prefix string, directories []*XmlDirectory, errs *strings.Builder) {
	for i := range directories {
		var directory = directories[i]
		var path = colljoin(prefix, directory.Name)
		coll.loadDirectories(path, directory.Directories, errs)
		coll.loadFiles(path, directory.Files, errs)
	}
}

func (coll *collection) loadFiles(prefix string, files []*XmlFile, errs *strings.Builder) {
	for i := range files {
		var xmlFile = files[i]
		var file, err = xmlFileToStd(xmlFile)
		if err != nil {
			errs.WriteString(err.Error()+"\n")
		} else {
			var collpath = colljoin(prefix, file.name)
			file.name = collpath
			coll.files[collpath] = file
		}
	}
}

func (coll *collection) StoreTo(out io.Writer) error {
	var xmlColl = new(XmlRsCollection)
	xmlColl.Files = make([]*XmlFile, len(coll.files))

	for _, v := range coll.files {
		var xmlFile = stdFileToXml(v)
		var path = collsplit(xmlFile.Name)
		if len(path) == 1 {
			xmlColl.Files = appendFileToSlice(xmlColl.Files, xmlFile)
		} else {
			var directory = findDirectory(xmlColl, path[:len(path)-1])
			xmlFile.Name = path[len(path)-1]
			directory.Files = appendFileToSlice(directory.Files, xmlFile)
		}
	}

	var _, err = out.Write([]byte(xmlHeader))

	var data []byte
	data, err = xml.MarshalIndent(xmlColl, "", "\t")

	if err == nil {
		_, err = out.Write(data)
	}

	return err
}

func findDirectory(xmlColl *XmlRsCollection, path [] string) *XmlDirectory {
	var dirs = xmlColl.Directories
	var firstName = path[0]
	var dir = findDirectoryInSlice(dirs, firstName)
	if dir == nil {
		xmlColl.Directories, dir = appendDirectoryToSlice(dirs, firstName)
	}

	path = path[1:]

	for i := range path {
		var p = path[i]
		var nextDir *XmlDirectory = findDirectoryInSlice(dir.Directories, p)
		if nextDir == nil {
			dir.Directories, nextDir = appendDirectoryToSlice(dir.Directories, p)
		}
		dir = nextDir
	}

	return dir
}

func findDirectoryInSlice(slice []*XmlDirectory, name string) *XmlDirectory {
	if slice == nil {
		return nil
	}
	
	for i := range slice {
		var dir = slice[i]
		if dir.Name == name {
			return dir
		}
	}
	return nil
}

func appendDirectoryToSlice(slice []*XmlDirectory, name string) ([]*XmlDirectory, *XmlDirectory) {
	var dir = new(XmlDirectory)
	dir.Name = name
	var ret = slice
	if ret == nil {
		ret = make([]*XmlDirectory, 4)[0:0]
	}
	var l = len(ret)
	if l == cap(ret) {
		var newSlice = make([]*XmlDirectory, l*2)
		copy(newSlice, ret)
		ret = newSlice
	}
	ret = ret[0:l+1]
	ret[l] = dir
	return ret, dir
}

func appendFileToSlice(slice []*XmlFile, file *XmlFile) []*XmlFile {
	var ret = slice
	if ret == nil {
		ret = make([]*XmlFile, 16)[0:0]
	}
	var l = len(ret)
	if l==cap(ret) {
		var newSlice = make([]*XmlFile, l*2)
		copy(newSlice, ret)
		ret = newSlice
	}
	ret = ret[0:l+1]
	ret[l] = file
	return ret
}

func xmlFileToStd(xmlFile *XmlFile) (*file, error) {
	var ret = new(file)
	ret.name = xmlFile.Name
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
	ret.Name = file.Name()
	ret.Sha1 = file.Sha1()
	ret.Size = file.Size()
	var timestamp = file.Timestamp()
	if timestamp != (time.Time{}) {
		ret.Updated = timestamp.Format(time.RFC3339Nano)
	}
	return ret
}

func (file *file) Name() string {
	return file.name
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

func colljoin(parent, child string) string {
	if parent == "" {
		return child
	} else {
		return parent+"/"+child
	}
}

func collsplit(path string) []string {
	return strings.Split(path, "/")
}
