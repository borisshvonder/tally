package tallylib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_UpdateSingleDirectory_1_file(t *testing.T) {
	dir, err := ioutil.TempDir("", "Test_UpdateSingleDirectory_1_file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	subdir := filepath.Join(dir, "subdir")
	os.Mkdir(subdir, os.ModeDir|os.ModePerm)

	tempfile := filepath.Join(subdir, "file1")
	ioutil.WriteFile(tempfile, []byte("Hello, world!"), os.ModePerm)

	fixture := NewTally()
	// Uncomment to enable logging:
	//fixture.SetLog(os.Stdout)
	var config TallyConfig
	config.logVerbosity = 100
	fixture.SetConfig(config)

	changed, err := fixture.UpdateSingleDirectory(subdir)
	if err != nil {
		t.Log("Cannot update", err)
		t.Fail()
	}
	if !changed {
		t.Log("tally did not report collection changed")
		t.Fail()
	}

	var coll = NewCollection()
	collFile, err := os.Open(filepath.Join(dir, "subdir.rscollection"))
	defer collFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = coll.LoadFrom(collFile)
	if err != nil {
		t.Fatal(err)
	}

	file := coll.ByName("file1")
	if file == nil {
		t.Log("no file1 found in collection")
		t.Fail()
	} else if "943a702d06f34599aee1f8da8ef9f7296031d699" != file.Sha1() {
		t.Log("file1 sha1 is " + file.Sha1())
		t.Fail()
	}
}
