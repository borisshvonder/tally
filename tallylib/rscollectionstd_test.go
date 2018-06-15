package tallylib

import (
	"testing"
	"time"
)

func TestRSCollectionFileStd(t *testing.T) {
	var now = time.Now()
	var file = new(RSCollectionFileStd)
	file.path = "path1"
	file.sha1 = "sha1"
	file.timestamp = now

	if file.Path() != "path1" {
		t.Log("Path() does not work")
		t.Fail()
	}

	if file.Sha1() != "sha1" {
		t.Log("Sha1() does not work")
		t.Fail()
	}

	if file.Timestamp() != now {
		t.Log("Timestamp() does not work")
		t.Fail()
	}
}
