package tallylib

import (
	"testing"
)

func TestUpdateFile_file_does_not_exist(t *testing.T) {
	var coll = New()
	coll.InitEmpty()

	var ret, err = UpdateFile(coll, "/path/does/notexist", "notexist", false)
	if err == nil {
		t.Log("Should fail on file that does not exist")
		t.Fail()
	}
	if ret {
		t.Log("Should not return true")
		t.Fail()
	}
}
