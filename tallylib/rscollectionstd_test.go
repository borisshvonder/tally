package tallylib

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRSCollectionFileStd(t *testing.T) {
	var now = time.Now()
	var fixture = new(RSCollectionFileStd)
	fixture.path = "path1"
	fixture.sha1 = "sha1"
	fixture.timestamp = now

	assertFile(t, fixture, "path1", "sha1", now)

}

func TestRSCollectionStd_Update(t *testing.T) {
	var fixture = new(RSCollectionStd)
	fixture.InitEmpty()

	var now = time.Now()
	var path = "relative/path"
	var sha1 = "sha1"

	var file = fixture.Update(path, sha1, now)
	assertFile(t, file, path, sha1, now)

	var fileByPath = fixture.ByPath(path)
	assertFile(t, fileByPath, path, sha1, now)
}

func TestRSCollectionStd_ByPath_nil(t *testing.T) {
	var fixture = new(RSCollectionStd)
	fixture.InitEmpty()

	var notExisting = fixture.ByPath("/missing/path")
	if notExisting != nil {
		t.Log("Expected nil, got " + notExisting.Path())
		t.Fail()
	}
}

func TestRSCollectionStd_Visit(t *testing.T) {
	var fixture = new(RSCollectionStd)
	fixture.InitEmpty()

	var now = time.Now()
	var path = "relative/path"
	var sha1 = "sha1"

	fixture.Update(path, sha1, now)

	var files [1]RSCollectionFile
	var i = 0
	fixture.Visit(func(file RSCollectionFile) {
		files[i] = file
		i++
	})

	if i != 1 {
		t.Log("got " + strconv.Itoa(i) + " files")
		t.Fail()
	}
	assertFile(t, files[0], path, sha1, now)
}

func TestRSCollectionStd_LoadFrom_empty_xml(t *testing.T) {
	var reader = strings.NewReader("")
	var fixture = new(RSCollectionStd)

	fixture.LoadFrom(reader)

	assertEmptyCollection(t, fixture)
}

func assertEmptyCollection(t *testing.T, fixture RSCollection) {
	fixture.Visit(func(file RSCollectionFile) {
		t.Log("Error: collection contains file " + file.Path())
		t.Fail()
	})
}

func assertFile(
	t *testing.T,
	file RSCollectionFile,
	path string,
	sha1 string,
	timestamp time.Time) {

	assertStrEquals(t, "file.Path()", path, file.Path())
	assertStrEquals(t, "file.Sha1()", sha1, file.Sha1())
	assertTimeEquals(t, "file.Timestamp()", timestamp, file.Timestamp())

}

func assertStrEquals(t *testing.T, context string, expected string, actual string) {

	if expected != actual {
		t.Log(context + ": expected '" + expected +
			"', actual '" + actual + "'")
		t.Fail()
	}
}

func assertTimeEquals(t *testing.T, context string, expected time.Time, actual time.Time) {

	if expected != actual {
		t.Log(context + ": expected '" + expected.String() +
			"', actual '" + actual.String() + "'")
		t.Fail()
	}
}
