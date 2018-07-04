package tallylib

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestUpdateFile_file_does_not_exist(t *testing.T) {
	var coll = NewCollection()
	coll.InitEmpty()

	var ret, err = updateFile(coll, "/path/does/notexist", "notexist", false)
	if err == nil {
		t.Log("Should fail on file that does not exist")
		t.Fail()
	}
	if ret {
		t.Log("Should not return true")
		t.Fail()
	}
}

func TestUpdate_existingFile(t *testing.T) {
	var temp, err = ioutil.TempFile("", "TestUpdate_existingFile")
	if err != nil {
		t.Fatal(err)
	}
	var path = temp.Name()
	defer os.Remove(path)
	defer temp.Close()

	temp.WriteString("hello")
	temp.Close()

	var coll = NewCollection()
	coll.InitEmpty()

	var ret bool
	ret, err = updateFile(coll, path, path, false)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !ret {
		t.Log("Failed to update file " + path)
		t.Fail()
	}
	if coll.ByPath(path) == nil {
		t.Log("Collection was not updated for file " + path)
		t.Fail()
	}

	ret, err = updateFile(coll, path, path, false)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if ret {
		t.Log("Collection WAS updated while it should not be")
		t.Fail()
	}

	ret, err = updateFile(coll, path, path, true)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !ret {
		t.Log("Collection was not updated even with force=true")
		t.Fail()
	}

	ioutil.WriteFile(path, []byte("Hello, again!"), os.ModePerm)

	ret, err = updateFile(coll, path, path, false)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !ret {
		t.Log("Collection was not updated even when file changed")
		t.Fail()
	}
}
