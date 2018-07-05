package tallylib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_UpdateSingleDirectory_1_file(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateSingleDirectory_1_file")
	defer os.RemoveAll(tmpdir)

	subdir := mkdir(tmpdir, "subdir")
	writefile(subdir, "file1", "Hello, world!")

	var coll = assertUpdateSingleDirectory(t, fixture, subdir)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
	assertCollectionSize(t, 1, coll)
}

func Test_UpdateSingleDirectory_2_files(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateSingleDirectory_2_files")
	defer os.RemoveAll(tmpdir)

	subdir1 := mkdir(tmpdir, "subdir1")
	writefile(subdir1, "file1", "Hello, world!")
	writefile(subdir1, "file2", "Hello again")
	mkdir(subdir1, "subdir2")

	var coll = assertUpdateSingleDirectory(t, fixture, subdir1)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
	assertFileInCollection(t, coll, "file2", "43ce0c8e7e28680735241ad3e5550aa361b96f53")
	assertCollectionSize(t, 2, coll)
}

func assertCollectionSize(t *testing.T, expected int, coll RSCollection) {
	actual := 0
	coll.Visit(func(f RSCollectionFile) {
		actual += 1
	})

	if expected != actual {
		t.Log("Collection size expected", expected, "but actual", actual)
		t.Fail()
	}
}

func assertFileInCollection(t *testing.T, coll RSCollection, name, sha1 string) {
	file := coll.ByName(name)
	if file == nil {
		t.Log("no", name, "found in collection")
		t.Fail()
	} else if sha1 != file.Sha1() {
		t.Log(name, "sha1 is", file.Sha1())
		t.Fail()
	}
}

func assertUpdateSingleDirectory(t *testing.T, fixture Tally, directory string) RSCollection {
	changed, err := fixture.UpdateSingleDirectory(directory)
	if err != nil {
		t.Log("Cannot update", err)
		t.Fail()
	}
	if !changed {
		t.Log("tally did not report collection changed")
		t.Fail()
	}
	parent := filepath.Dir(directory)
	name := filepath.Base(directory)
	collFile := filepath.Join(parent, name+".rscollection")
	return loadCollection(t, collFile)
}

func loadCollection(t *testing.T, filepath string) RSCollection {
	var coll = NewCollection()
	collFile, err := os.Open(filepath)
	defer collFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = coll.LoadFrom(collFile)
	if err != nil {
		t.Fatal(err)
	}

	return coll
}

func createFixture() Tally {
	fixture := NewTally()
	// Uncomment to enable logging:
	fixture.SetLog(os.Stdout)
	var config TallyConfig
	config.logVerbosity = 100
	fixture.SetConfig(config)
	return fixture
}

func mktmp(prefix string) string {
	tmpdir, err := ioutil.TempDir("", prefix)
	if err != nil {
		panic(err)
	}
	return tmpdir
}

func mkdir(parent, name string) string {
	ret := filepath.Join(parent, name)
	err := os.Mkdir(ret, os.ModeDir|os.ModePerm)
	if err != nil {
		panic(err)
	}
	return ret
}

func writefile(parent, name, contents string) string {
	ret := filepath.Join(parent, name)
	err := ioutil.WriteFile(ret, []byte(contents), os.ModePerm)
	if err != nil {
		panic(err)
	}
	return ret
}
