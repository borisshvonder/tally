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
	var fixture = NewCollection()
	fixture.InitEmpty()

	var now = time.Now()
	var name = "relative/path"
	var size int64 = 1234
	var sha1 = "sha1"

	var file = fixture.Update(name, sha1, size, now)
	assertFile(t, file, name, sha1, size, now)

	var fileByName = fixture.ByName(name)
	assertFile(t, fileByName, name, sha1, size, now)
	assertIntEquals(t, "coll.Size()", 1, fixture.Size())
}

func Test_Remove(t *testing.T) {
	var fixture = NewCollection()
	fixture.InitEmpty()

	fixture.Update("file1", "sha1", 0, time.Now())
	fixture.Update("file2", "sha2", 0, time.Now())

	fixture.RemoveFile("file1")
	if fixture.ByName("file1") != nil {
		t.Log("file1 was not removed")
		t.Fail()
	}
	if fixture.ByName("file2") == nil {
		t.Log("file2 WAS removed")
		t.Fail()
	}
	assertIntEquals(t, "coll.Size()", 1, fixture.Size())
}

func Test_UpdateFile(t *testing.T) {
	var source = NewCollection()
	source.InitEmpty()

	var now = time.Now()
	var name = "relative/path"
	var size int64 = 1234
	var sha1 = "sha1"

	var file = source.Update(name, sha1, size, now)

	var fixture = NewCollection()
	fixture.InitEmpty()

	fixture.UpdateFile(file)
	file = fixture.ByName(name)
	assertFile(t, file, name, sha1, size, now)
}

func Test_ByName_nil(t *testing.T) {
	var fixture = NewCollection()
	fixture.InitEmpty()

	var notExisting = fixture.ByName("missing/path")
	if notExisting != nil {
		t.Log("Expected nil, got " + notExisting.Name())
		t.Fail()
	}
}

func Test_Visit(t *testing.T) {
	var fixture = NewCollection()
	fixture.InitEmpty()

	var now = time.Now()
	var name = "relative/path"
	var size int64 = 12345
	var sha1 = "sha1"

	fixture.Update(name, sha1, size, now)

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
	assertFile(t, files[0], name, sha1, size, now)
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
		var file = fixture.ByName("name1")

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
		assertFile(t, fixture.ByName("name1"), "name1",
			"8551d11f6e8d3ec2731f70a2573b887637e94559",
			1024, (time.Time{}))

		assertFile(t, fixture.ByName("name2"), "name2",
			"6551d11f6e8d3ec2731f70a2573b887637e94559",
			0, (time.Time{}))
	}
}


func Test_LoadFrom_subdirectories(t *testing.T) {
	var xml = `<!DOCTYPE RsCollection>
		<RsCollection>
			<Directory name="subdir1">
				<Directory name="subdir2">
					<File sha1="6551d11f6e8d3ec2731f70a2573b887637e94559" 
						name="name2" />
				</Directory>
				<File sha1="8551d11f6e8d3ec2731f70a2573b887637e94559" 
					name="name1" size="1024" />
			</Directory>
		</RsCollection>`

	var fixture, err = loadCollectionFromString(xml)
	if err != nil {
		failOnError(t, err)
	} else {
		assertFile(t, fixture.ByName("subdir1/name1"), "subdir1/name1",
			"8551d11f6e8d3ec2731f70a2573b887637e94559",
			1024, (time.Time{}))

		assertFile(t, fixture.ByName("subdir1/subdir2/name2"), "subdir1/subdir2/name2",
			"6551d11f6e8d3ec2731f70a2573b887637e94559",
			0, (time.Time{}))
	}
}


func Test_StoreTo_emptyCollection(t *testing.T) {
	var coll = NewCollection()
	var buf bytes.Buffer
	coll.StoreTo(io.Writer(&buf))

	var str = buf.String()
	assertStrEquals(t, "coll.StoreTo()", "<!DOCTYPE RsCollection>\n<RsCollection></RsCollection>", str)
}

func Test_StoreTo_oneRecord(t *testing.T) {
	var coll = NewCollection()
	coll.InitEmpty()
	coll.Update("name1", "sha1", 1024, (time.Time{}))

	var buf bytes.Buffer
	coll.StoreTo(io.Writer(&buf))

	var str = buf.String()
	assertStrEquals(t, "coll.StoreTo()",
		`<!DOCTYPE RsCollection>
<RsCollection>
	<File sha1="sha1" name="name1" size="1024" updated=""></File>
</RsCollection>`, str)
}

func Test_StoreTo_multipleRecords(t *testing.T) {
	var coll = NewCollection()
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


func Test_StoreTo_Subdirectories(t *testing.T) {
	var coll = NewCollection()
	coll.InitEmpty()
	coll.Update("name0", "sha0", 10240, (time.Time{}))
	coll.Update("dir1/name1", "sha1", 10241, (time.Time{}))
	coll.Update("dir1/dir2/name2", "sha2", 10242, (time.Time{}))
	coll.Update("dir1/dir2/dir3/name3", "sha3", 10243, (time.Time{}))

	var buf bytes.Buffer
	coll.StoreTo(io.Writer(&buf))

	var str = buf.String()
	assertStrEquals(t, "coll.StoreTo()",
		`<!DOCTYPE RsCollection>
<RsCollection>
	<Directory name="dir1">
		<Directory name="dir2">
			<Directory name="dir3">
				<File sha1="sha3" name="name3" size="10243" updated=""></File>
			</Directory>
			<File sha1="sha2" name="name2" size="10242" updated=""></File>
		</Directory>
		<File sha1="sha1" name="name1" size="10241" updated=""></File>
	</Directory>
	<File sha1="sha0" name="name0" size="10240" updated=""></File>
</RsCollection>`, str)

}

func failOnError(t *testing.T, err error) {
	t.Log(err.Error())
	t.Fail()
}

func loadCollectionFromString(xml string) (RSCollection, error) {
	var reader = strings.NewReader(xml)
	var fixture = NewCollection()
	var err = fixture.LoadFrom(reader)
	return fixture, err
}

func assertEmptyCollection(t *testing.T, fixture RSCollection) {
	fixture.Visit(func(file RSCollectionFile) {
		t.Log("Error: collection contains file " + file.Name())
		t.Fail()
	})
}

func assertFile(t *testing.T, file RSCollectionFile, name, sha1 string, size int64, timestamp time.Time) {
	if file == nil {
		t.Log("file", name, "not found in collection")
		t.Fail()
	} else {
		assertStrEquals(t, "file.Name()", name, file.Name())
		assertStrEquals(t, "file.Sha1()", sha1, file.Sha1())
		assertUint64Equals(t, "file.Size()", size, file.Size())
		assertTimeEquals(t, "file.Timestamp()", timestamp, file.Timestamp())
	}
}

func assertStrEquals(t *testing.T, context, expected, actual string) {

	if expected != actual {
		t.Log(context + ": expected '" + expected +
			"', actual '" + actual + "'")
		t.Fail()
	}
}

func assertIntEquals(t *testing.T, context string, expected, actual int) {
	if expected != actual {
		t.Log(context + ": expected '" +
			strconv.Itoa(expected) +
			"', actual '" +
			strconv.Itoa(actual) +
			"'")
		t.Fail()
	}
}

func assertUint64Equals(t *testing.T, context string, expected, actual int64) {
	if expected != actual {
		t.Log(context + ": expected '" +
			strconv.FormatInt(expected, 10) +
			"', actual '" +
			strconv.FormatInt(actual, 10) +
			"'")
		t.Fail()
	}
}

func assertTimeEquals(t *testing.T, context string, expected, actual time.Time) {
	if expected != actual {
		t.Log(context + ": expected '" + expected.String() +
			"', actual '" + actual.String() + "'")
		t.Fail()
	}
}
