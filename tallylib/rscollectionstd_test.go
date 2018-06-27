package tallylib

import (
	"bytes"
	"io"
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
	var size uint64 = 1234
	var sha1 = "sha1"

	var file = fixture.Update(path, sha1, size, now)
	assertFile(t, file, path, sha1, size, now)

	var fileByPath = fixture.ByPath(path)
	assertFile(t, fileByPath, path, sha1, size, now)
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
	var size uint64 = 12345
	var sha1 = "sha1"

	fixture.Update(path, sha1, size, now)

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
	assertFile(t, files[0], path, sha1, size, now)
}

func Test_LoadFrom_empty_xml(t *testing.T) {
	var fixture, err = loadCollectionFromString("")
	if err == nil {
		t.Log("EOF error expected on empty file")
		t.Fail()
	} else {
		assertEmptyCollection(t, fixture)
	}
}

func Test_LoadFrom_empty(t *testing.T) {
	var xml = "<RsCollection/>"

	var fixture, err = loadCollectionFromString(xml)
	if err != nil {
		failOnError(t, err)
	} else {
		var size = 0
		fixture.Visit(func(file RSCollectionFile) {
			size += 1
		})

		assertIntEquals(t, "Test_LoadFrom_empty", 0, size)
	}
}

func Test_LoadFrom_oneRecord(t *testing.T) {
	var xml = `<!DOCTYPE RsCollection>
		<RsCollection>
			<File sha1="8551d11f6e8d3ec2731f70a2573b887637e94559" 
				name="name1" size="1024" 
       				updated="2018-02-28T18:30:01.123Z"/>
		</RsCollection>`

	var fixture, err = loadCollectionFromString(xml)
	if err != nil {
		failOnError(t, err)
	} else {
		var file = fixture.ByPath("name1")

		var expectedTimestamp, err = time.Parse(time.RFC3339,
			"2018-02-28T18:30:01.123Z")

		if err != nil {
			failOnError(t, err)
		}

		assertFile(t, file, "name1",
			"8551d11f6e8d3ec2731f70a2573b887637e94559",
			1024, expectedTimestamp)
	}
}

func Test_LoadFrom_multipleRecords(t *testing.T) {
	var xml = `<!DOCTYPE RsCollection>
		<RsCollection>
			<File sha1="8551d11f6e8d3ec2731f70a2573b887637e94559" 
				name="name1" size="1024" />
			<File sha1="6551d11f6e8d3ec2731f70a2573b887637e94559" 
				name="name2" />
		</RsCollection>`

	var fixture, err = loadCollectionFromString(xml)
	if err != nil {
		failOnError(t, err)
	} else {
		assertFile(t, fixture.ByPath("name1"), "name1",
			"8551d11f6e8d3ec2731f70a2573b887637e94559",
			1024, (time.Time{}))

		assertFile(t, fixture.ByPath("name2"), "name2",
			"6551d11f6e8d3ec2731f70a2573b887637e94559",
			0, (time.Time{}))
	}
}

func Test_StoreTo_emptyCollection(t *testing.T) {
	var coll = New()
	var buf bytes.Buffer
	coll.StoreTo(io.Writer(&buf))

	var str = buf.String()
	assertStrEquals(t, "coll.StoreTo()", "<!DOCTYPE RsCollection>\n<RsCollection></RsCollection>", str)
}

func Test_StoreTo_oneRecord(t *testing.T) {
	var coll = New()
	coll.InitEmpty()
	coll.Update("name1", "sha1", 1024, (time.Time{}))

	var buf bytes.Buffer
	coll.StoreTo(io.Writer(&buf))

	var str = buf.String()
	assertStrEquals(t, "coll.StoreTo()",
		`<!DOCTYPE RsCollection>
<RsCollection><File sha1="sha1" name="name1" size="1024" updated=""></File></RsCollection>`, str)
}

func Test_StoreTo_multipleRecords(t *testing.T) {
	var coll = New()
	coll.InitEmpty()
	coll.Update("name1", "sha1", 0, (time.Time{}))
	coll.Update("name2", "", 0, (time.Time{}))

	var buf bytes.Buffer
	coll.StoreTo(io.Writer(&buf))

	var str = buf.String()

	if strings.Index(str, "name1") < 0 {
		t.Log("name1 not found in " + str)
		t.Fail()
	}

	if strings.Index(str, "sha1") < 0 {
		t.Log("sha1 not found in " + str)
		t.Fail()
	}
	if strings.Index(str, "name2") < 0 {
		t.Log("name2 not found in " + str)
		t.Fail()
	}

}

func failOnError(t *testing.T, err error) {
	t.Log(err.Error())
	t.Fail()
}

func loadCollectionFromString(xml string) (RSCollection, error) {
	var reader = strings.NewReader(xml)
	var fixture = New()
	var err = fixture.LoadFrom(reader)
	return fixture, err
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
	size uint64,
	timestamp time.Time) {

	assertStrEquals(t, "file.Path()", path, file.Path())
	assertStrEquals(t, "file.Sha1()", sha1, file.Sha1())
	assertUint64Equals(t, "file.Size()", size, file.Size())
	assertTimeEquals(t, "file.Timestamp()", timestamp, file.Timestamp())
}

func assertStrEquals(t *testing.T, context string, expected string,
	actual string) {

	if expected != actual {
		t.Log(context + ": expected '" + expected +
			"', actual '" + actual + "'")
		t.Fail()
	}
}

func assertIntEquals(t *testing.T, context string, expected int,
	actual int) {

	if expected != actual {
		t.Log(context + ": expected '" +
			strconv.Itoa(expected) +
			"', actual '" +
			strconv.Itoa(actual) +
			"'")
		t.Fail()
	}
}

func assertUint64Equals(t *testing.T, context string, expected uint64,
	actual uint64) {

	if expected != actual {
		t.Log(context + ": expected '" +
			strconv.FormatUint(expected, 10) +
			"', actual '" +
			strconv.FormatUint(actual, 10) +
			"'")
		t.Fail()
	}
}

func assertTimeEquals(t *testing.T, context string, expected time.Time,
	actual time.Time) {

	if expected != actual {
		t.Log(context + ": expected '" + expected.String() +
			"', actual '" + actual.String() + "'")
		t.Fail()
	}
}
