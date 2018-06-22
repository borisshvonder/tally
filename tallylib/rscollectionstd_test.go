package tallylib

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_Update(t *testing.T) {
	var fixture = New()
	fixture.InitEmpty()

	var now = time.Now()
	var path = "relative/path"
	var sha1 = "sha1"

	var file = fixture.Update(path, sha1, now)
	assertFile(t, file, path, sha1, now)

	var fileByPath = fixture.ByPath(path)
	assertFile(t, fileByPath, path, sha1, now)
}

func Test_ByPath_nil(t *testing.T) {
	var fixture = New()
	fixture.InitEmpty()

	var notExisting = fixture.ByPath("/missing/path")
	if notExisting != nil {
		t.Log("Expected nil, got " + notExisting.Path())
		t.Fail()
	}
}

func Test_Visit(t *testing.T) {
	var fixture = New()
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

func Test_LoadFrom_empty_xml(t *testing.T) {
	var fixture = loadCollectionFromString("")

	assertEmptyCollection(t, fixture)
}

func Test_LoadFrom_oneRecord(t *testing.T) {
	var xml = `<!DOCTYPE RsCollection>
		<RsCollection>
			<File sha1="8551d11f6e8d3ec2731f70a2573b887637e94559" 
				name="name1" size="1024" 
       				updated="2018-02-28T18:30:01.123Z"/>
		</RsCollection>`

	var fixture = loadCollectionFromString(xml)

	var file = fixture.ByPath("name1")
	var expectedTimestamp, _ = time.Parse(time.RFC3339, "2018-02-28T18:30:01.123Z")

	assertFile(t, file, "name1", "8551d11f6e8d3ec2731f70a2573b887637e94559",
		expectedTimestamp)
}

func loadCollectionFromString(xml string) RSCollection {
	var reader = strings.NewReader(xml)
	var fixture = New()
	fixture.LoadFrom(reader)
	return fixture
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
