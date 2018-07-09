package tallylib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createFixture() Tally {
	fixture := NewTally()
	// Uncomment to enable logging:
	// fixture.SetLog(os.Stdout)
	var config TallyConfig
	config.logVerbosity = 100
	fixture.SetConfig(config)
	return fixture
}

func Test_DefaultConfig(t *testing.T) {
	fixture := NewTally()
	config := fixture.GetConfig()
	if config.forceUpdate || config.stopOnWarnings || config.logVerbosity != 3 {
		t.Log("Invalid default config")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_fail_when_no_directory(t *testing.T) {
	fixture := createFixture()
	
	var _, err = fixture.UpdateSingleDirectory("this-directory-does-notexist")
	if err == nil {
		t.Log("Should fail when directory does not exist")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_fail_when_no_access(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateSingleDirectory_will_fail_when_no_access")
	defer os.RemoveAll(tmpdir)

	subdir := filepath.Join(tmpdir, "forbidden")
	os.Mkdir(subdir, 0)
	
	var _, err = fixture.UpdateSingleDirectory(subdir)
	if err == nil {
		t.Log("Should fail when directory has incorrect permssions")
		t.Fail()
	}
}


func Test_UpdateSingleDirectory_will_fail_when_pointed_to_file(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateSingleDirectory_will_fail_when_pointed_to_file")
	defer os.RemoveAll(tmpdir)

	subdir := mkdir(tmpdir, "forbidden")
	var file =  writefile(subdir, "file1", "Hello, world!")
	
	var _, err = fixture.UpdateSingleDirectory(file)
	if err == nil {
		t.Log("Should fail when target directory is a file")
		t.Fail()
	}
}


func Test_UpdateSingleDirectory_will_fail_when_no_access_to_file(t *testing.T) {
	var fixture = createFixture()
	var config = fixture.GetConfig()
	config.stopOnWarnings = true
	fixture.SetConfig(config)

	var tmpdir = mktmp("Test_UpdateSingleDirectory_fail_when_no_access_to_file")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	var file =  writefile(subdir, "file1", "Hello, world!")
	os.Chmod(file, 0)
	
	var _, err = fixture.UpdateSingleDirectory(subdir)
	if err == nil {
		t.Log("Should fail when input file has no read permssions")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_ignore_no_access_to_file(t *testing.T) {
	var fixture = createFixture()

	var tmpdir = mktmp("Test_UpdateSingleDirectory_will_fail_when_invalid_coecton_file")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	writefile(subdir, "file1", "Hello, world!")
	var file2 = writefile(subdir, "file2", "Hello, world!")
	os.Chmod(file2, 0)

	var coll = assertUpdateSingleDirectory(t, fixture, subdir)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
	assertCollectionSize(t, 1, coll)
	mkdir(subdir, "this-should-be-ignored")
	assertWillNotUpdateSingleDirectory(t, fixture, subdir)
}


func Test_UpdateSingleDirectory_will_fail_when_invalid_collecton_file(t *testing.T) {
	var fixture = createFixture()
	var config = fixture.GetConfig()
	config.stopOnWarnings = true
	fixture.SetConfig(config)

	var tmpdir = mktmp("Test_UpdateSingleDirectory_will_fail_when_invalid_coecton_file")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	writefile(tmpdir, "subdir.rscollection", "INVALID")
	
	var _, err = fixture.UpdateSingleDirectory(subdir)
	if err == nil {
		t.Log("Should fail when stopOnWarnings and collection file is bad")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_not_fail_when_no_access_to_collecton_file(t *testing.T) {
	var fixture = createFixture()
	var tmpdir = mktmp("Test_UpdateSingleDirectory_will_not_fail_when_no_access_to_collecton_file")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	var collectionFile =  writefile(tmpdir, "subdir.rscollection", "<RsCollection/>")
	os.Chmod(collectionFile, 0)
	
	var _, err = fixture.UpdateSingleDirectory(subdir)
	if err == nil {
		t.Log("Should fail when input file has no read permssions")
		t.Fail()
	}
}




func Test_UpdateSingleDirectory(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateSingleDirectory")
	defer os.RemoveAll(tmpdir)
	
	subdir := mkdir(tmpdir, "subdir")
	writefile(subdir, "file1", "Hello, world!")
	var coll = assertUpdateSingleDirectory(t, fixture, subdir)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
	assertCollectionSize(t, 1, coll)
	mkdir(subdir, "this-should-be-ignored")
	assertWillNotUpdateSingleDirectory(t, fixture, subdir)


	writefile(subdir, "file2", "Hello again")
	coll = assertUpdateSingleDirectory(t, fixture, subdir)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
	assertFileInCollection(t, coll, "file2", "43ce0c8e7e28680735241ad3e5550aa361b96f53")
	assertCollectionSize(t, 2, coll)
	assertWillNotUpdateSingleDirectory(t, fixture, subdir)

	writefile(subdir, "file3", "And again")
	coll = assertUpdateSingleDirectory(t, fixture, subdir)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
	assertFileInCollection(t, coll, "file2", "43ce0c8e7e28680735241ad3e5550aa361b96f53")
	assertFileInCollection(t, coll, "file3", "b32642de88c24a48f9de7f76698eb7a9a65dae58")
	assertCollectionSize(t, 3, coll)
	assertWillNotUpdateSingleDirectory(t, fixture, subdir)

	writefile(subdir, "file1", "Change contents")
	coll = assertUpdateSingleDirectory(t, fixture, subdir)
	assertFileInCollection(t, coll, "file1", "8804b55dbfd918a5bd47cf31e8f5ec8ccae6abb7")
	assertFileInCollection(t, coll, "file2", "43ce0c8e7e28680735241ad3e5550aa361b96f53")
	assertFileInCollection(t, coll, "file3", "b32642de88c24a48f9de7f76698eb7a9a65dae58")
	assertCollectionSize(t, 3, coll)
	assertWillNotUpdateSingleDirectory(t, fixture, subdir)

	writefile(tmpdir, "subdir.rscollection", "badxml")
	coll = assertUpdateSingleDirectory(t, fixture, subdir)
	assertFileInCollection(t, coll, "file1", "8804b55dbfd918a5bd47cf31e8f5ec8ccae6abb7")
	assertFileInCollection(t, coll, "file2", "43ce0c8e7e28680735241ad3e5550aa361b96f53")
	assertFileInCollection(t, coll, "file3", "b32642de88c24a48f9de7f76698eb7a9a65dae58")
	assertCollectionSize(t, 3, coll)
	assertWillNotUpdateSingleDirectory(t, fixture, subdir)

	os.Remove(filepath.Join(subdir, "file2"))
	coll = assertUpdateSingleDirectory(t, fixture, subdir)
	assertFileInCollection(t, coll, "file1", "8804b55dbfd918a5bd47cf31e8f5ec8ccae6abb7")
	assertFileInCollection(t, coll, "file3", "b32642de88c24a48f9de7f76698eb7a9a65dae58")
	assertCollectionSize(t, 2, coll)
	assertWillNotUpdateSingleDirectory(t, fixture, subdir)
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
func assertWillNotUpdateSingleDirectory(t *testing.T, fixture Tally, directory string) {
	var collectionFile = resolveCollectionFileForDirectory(directory)
	var oldTimestamp = getTimestampSafe(collectionFile)
	if update(t, fixture, directory) {
		t.Log("tally reported collection changed while it was not supposed to")
		t.Fail()
	}
	var newTimestamp = getTimestampSafe(collectionFile)
	if oldTimestamp != newTimestamp {
		t.Log("tally should NOT touch collection file!")
	}
}

func assertUpdateSingleDirectory(t *testing.T, fixture Tally, directory string) RSCollection {
	if !update(t, fixture, directory) {
		t.Log("tally did not report collection changed")
		t.Fail()
	}
	return loadCollectionForDirectory(t, directory)
}

func update(t *testing.T, fixture Tally, directory string) bool {
	changed, err := fixture.UpdateSingleDirectory(directory)
	if err != nil {
		t.Log("Cannot update", err)
		t.Fail()
	}
	return changed
}

func loadCollectionForDirectory(t *testing.T, directory string) RSCollection {
	collectionFile := resolveCollectionFileForDirectory(directory)
	return loadCollection(t, collectionFile)
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

func getTimestampSafe(filepath string) time.Time {
	stat, err := os.Stat(filepath)
	if err != nil {
		return time.Time{}
	} else {
		return stat.ModTime()
	}
}

func assertStringEquals(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Log("Expected:", expected, "actual:", actual)
		t.Fail()
	}
}

func Test_resolveCollectionFileForDirectory(t *testing.T) {
	assertStringEquals(t, "dir.rscollection", resolveCollectionFileForDirectory("dir"))
	assertStringEquals(t, "dir.rscollection", resolveCollectionFileForDirectory("dir/"))
	assertStringEquals(t, "parent/dir.rscollection", resolveCollectionFileForDirectory("parent/dir"))
	assertStringEquals(t, "/parent/dir.rscollection", resolveCollectionFileForDirectory("/parent/dir"))
	assertStringEquals(t, "/parent/dir.rscollection", resolveCollectionFileForDirectory("/parent/dir/"))
	assertStringEquals(t, ".rscollection", resolveCollectionFileForDirectory(""))
	assertStringEquals(t, "/.rscollection", resolveCollectionFileForDirectory("/"))
}

func Test_loadExistingCollection_will_fail_if_trying_to_load_from_dir(t *testing.T) {
	var fixture = new(tally)
	//fixture.config.logVerbosity = 100
	//fixture.SetLog(os.Stdout)

	var _, err = fixture.loadExistingCollection(".")
	if err == nil {
		t.Log("Should fail when trying to load from non-file")
		t.Fail()
	}
}
