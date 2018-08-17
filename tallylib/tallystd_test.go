package tallylib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
	"strings"
)

func createFixture() Tally {
	fixture := NewTally()
	// Uncomment to enable logging:
	//fixture.SetLog(os.Stdout)
	var config TallyConfig = fixture.GetConfig()
	config.LogVerbosity = 100
	fixture.SetConfig(config)
	return fixture
}

func Test_DefaultConfig(t *testing.T) {
	fixture := NewTally()
	config := fixture.GetConfig()
	if config.ForceUpdate || config.IgnoreWarnings || config.LogVerbosity != 3 {
		t.Log("Invalid default config")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_fail_when_no_directory(t *testing.T) {
	fixture := createFixture()
	
	var _, err = fixture.UpdateSingleDirectory("this-directory-does-notexist", false)
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
	
	var _, err = fixture.UpdateSingleDirectory(subdir, false)
	if err == nil {
		t.Log("Should fail when directory has incorrect permssions")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_fail_when_filename_expression_is_empty(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateSingleDirectory_will_fail_when_no_access")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	var config = fixture.GetConfig()
	config.CollectionPathnameExpression=""
	fixture.SetConfig(config)

	var _, err = fixture.UpdateSingleDirectory(subdir, false)
	if err == nil {
		t.Log("Should fail due to collection file expression empty")
		t.Fail()
	}
}


func Test_UpdateSingleDirectory_will_fail_when_filename_expression_is_incorrect(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateSingleDirectory_will_fail_when_no_access")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	var config = fixture.GetConfig()
	config.CollectionPathnameExpression="{{.Iinvalid call}}"
	fixture.SetConfig(config)

	var _, err = fixture.UpdateSingleDirectory(subdir, false)
	if err == nil {
		t.Log("Should fail due to collection file expression invalid")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_fail_when_pointed_to_file(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateSingleDirectory_will_fail_when_pointed_to_file")
	defer os.RemoveAll(tmpdir)

	subdir := mkdir(tmpdir, "forbidden")
	var file =  writefile(subdir, "file1", "Hello, world!")
	
	var _, err = fixture.UpdateSingleDirectory(file, false)
	if err == nil {
		t.Log("Should fail when target directory is a file")
		t.Fail()
	}
}


func Test_UpdateSingleDirectory_will_fail_when_no_access_to_file(t *testing.T) {
	var fixture = createFixture()

	var tmpdir = mktmp("Test_UpdateSingleDirectory_fail_when_no_access_to_file")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	var file =  writefile(subdir, "file1", "Hello, world!")
	os.Chmod(file, 0)
	
	var _, err = fixture.UpdateSingleDirectory(subdir, false)
	if err == nil {
		t.Log("Should fail when input file has no read permssions")
		t.Fail()
	}
}


func Test_UpdateSingleDirectory_will_ignore_no_access_to_file(t *testing.T) {
	var fixture = createFixture()
	var config = fixture.GetConfig()
	config.IgnoreWarnings = true
	fixture.SetConfig(config)

	var tmpdir = mktmp("Test_UpdateSingleDirectory_will_ignore_no_access_to_file")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	writefile(subdir, "file1", "Hello, world!")
	var file2 = writefile(subdir, "file2", "Hello, world!")
	os.Chmod(file2, 0111)

	var coll = assertUpdateSingleDirectory(t, fixture, subdir)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
	assertCollectionSize(t, 1, coll)
	mkdir(subdir, "this-should-be-ignored")
	assertWillNotUpdateSingleDirectory(t, fixture, subdir)
}


func Test_UpdateSingleDirectory_will_fail_when_invalid_collecton_file(t *testing.T) {
	var fixture = createFixture()

	var tmpdir = mktmp("Test_UpdateSingleDirectory_will_fail_when_invalid_coecton_file")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	writefile(tmpdir, "subdir.rscollection", "INVALID")
	
	var _, err = fixture.UpdateSingleDirectory(subdir, false)
	if err == nil {
		t.Log("Should fail when collection file is bad")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_not_fail_when_has_full_access_to_collecton_file(t *testing.T) {
	var fixture = createFixture()
	if !do_test_UpdateSingleDirectory_when_no_access_to_collecton_file(fixture, os.ModePerm) {
		t.Log("Should fail if no read access to collection file")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_fail_when_no_read_access_to_collecton_file(t *testing.T) {
	var fixture = createFixture()
	if do_test_UpdateSingleDirectory_when_no_access_to_collecton_file(fixture, 0) {
		t.Log("Should fail if no read access to collection file")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_fail_when_no_write_access_to_collecton_file(t *testing.T) {
	var fixture = createFixture()
	if do_test_UpdateSingleDirectory_when_no_access_to_collecton_file(fixture, 0555) {
		t.Log("Should fail if no read access to collection file")
		t.Fail()
	}
}

func do_test_UpdateSingleDirectory_when_no_access_to_collecton_file(fixture Tally, perm os.FileMode) bool {
	var tmpdir = mktmp("do_test_UpdateSingleDirectory_when_no_access_to_collecton_file")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	writefile(subdir, "file1", "change")
	var collectionFile =  writefile(tmpdir, "subdir.rscollection", "<RsCollection/>")
	os.Chmod(collectionFile, perm)
	
	var _, err = fixture.UpdateSingleDirectory(subdir, false)
	if err != nil {
		print(err)
	}
	return err == nil
}

func Test_will_fail_if_collectionFile_is_directory(t *testing.T) {
	var fixture = createFixture()
	var tmpdir = mktmp("Test_will_fail_if_collectionFile_is_directory")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	mkdir(tmpdir, "subdir.rscollection")

	var _, err = fixture.UpdateSingleDirectory(subdir, false)
	if err == nil {
		t.Log("Should fail when trying to load from non-file")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_not_removeExtra_files(t *testing.T) {
	fixture := createFixture()
	
	tmpdir := mktmp("Test_UpdateSingleDirectory_will_not_removeExtra_files")
	defer os.RemoveAll(tmpdir)
	
	subdir1 := mkdir(tmpdir, "subdir1")
	writefile(subdir1, "file1", "Hello, world!")

	subdir2 := mkdir(subdir1, "subdir2")
	writefile(subdir2, "file2", "Hello, world!")

	var coll2  = assertUpdateSingleDirectory(t, fixture, subdir2)
	var coll1  = assertUpdateSingleDirectory(t, fixture, subdir1)

	var file2 = coll2.ByName("file2")
	var relName = filepath.Join("subdir2", file2.Name())
	coll1.Update(relName, file2.Sha1(), file2.Size(), file2.Timestamp())
	var collFile, err = os.Create(filepath.Join(tmpdir, "subdir1.rscollection"))
	if err != nil {
		panic(err)
	}
	err = coll1.StoreTo(collFile)
	if err != nil {
		panic(err)
	}
	err = collFile.Close()
	if err != nil {
		panic(err)
	}

	assertWillNotUpdateSingleDirectory(t, fixture, subdir1)

	writefile(subdir1, "file1", "Change me")
	coll1 = assertUpdateSingleDirectory(t, fixture, subdir1)
	if coll1.ByName(relName) == nil {
		t.Log("File", relName, "gone from collection")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_will_RemoveExtraFiles(t *testing.T) {
	fixture := createFixture()
	var config = fixture.GetConfig()
	config.RemoveExtraFiles = true
	fixture.SetConfig(config)
	
	tmpdir := mktmp("Test_UpdateSingleDirectory_will_RemoveExtraFiles")
	defer os.RemoveAll(tmpdir)
	
	subdir1 := mkdir(tmpdir, "subdir1")
	writefile(subdir1, "file1", "Hello, world!")

	subdir2 := mkdir(subdir1, "subdir2")
	writefile(subdir2, "file2", "Hello, world!")

	var coll2  = assertUpdateSingleDirectory(t, fixture, subdir2)
	var coll1  = assertUpdateSingleDirectory(t, fixture, subdir1)

	var file2 = coll2.ByName("file2")
	var relName = filepath.Join("subdir2", file2.Name())
	coll1.Update(relName, file2.Sha1(), file2.Size(), file2.Timestamp())
	var collFile, err = os.Create(filepath.Join(tmpdir, "subdir1.rscollection"))
	if err != nil {
		panic(err)
	}
	err = coll1.StoreTo(collFile)
	if err != nil {
		panic(err)
	}
	err = collFile.Close()
	if err != nil {
		panic(err)
	}
	coll1 = assertUpdateSingleDirectory(t, fixture, subdir1)
	if coll1.ByName(relName) != nil {
		t.Log("File", relName, "should be removed")
		t.Fail()
	}
}

func Test_UpdateSingleDirectory_with_custom_collection_name_expression(t *testing.T) {
	fixture := createFixture()
	var config = fixture.GetConfig()
	config.IgnoreWarnings = true
	config.CollectionPathnameExpression = "{{.Path -1}}-{{.Path 0}}.rscollection"
	fixture.SetConfig(config)
	tmpdir := mktmp("Test_UpdateSingleDirectoryi_with_custom_collection_name_expression")
	defer os.RemoveAll(tmpdir)
	
	subdir1 := mkdir(tmpdir, "subdir1")
	subdir2 := mkdir(subdir1, "subdir2")
	writefile(subdir2, "file2", "Hello, world!")

	if !update(t, fixture, subdir2, false) {
		t.Log("tally did not report collection changed")
		t.Fail()
	}

	var collFile = filepath.Join(subdir1, "subdir1-subdir2.rscollection")
	var coll = loadCollection(t, collFile)
	assertFileInCollection(t, coll, "file2", "943a702d06f34599aee1f8da8ef9f7296031d699")
	assertCollectionSize(t, 1, coll)
}

func Test__UpdateSingleDirectoryi_with_custom_absolute_collection_name_expression(t *testing.T) {
        tmpdir := mktmp("Test_UpdateSingleDirectoryi_with_custom_collection_name_expression")
        defer os.RemoveAll(tmpdir)                                              
	colldir, err := filepath.Abs(mkdir(tmpdir, "collections"))
	if err != nil {
		panic("cannot resolve absolute path:"+err.Error())
	}
                                                                                
        fixture := createFixture()                                              
        var config = fixture.GetConfig()                                        
        config.IgnoreWarnings = true                                            
	
        config.CollectionPathnameExpression = filepath.Join(colldir, "{{.Path 0}}.rscollection")
        fixture.SetConfig(config)                                               

        subdir := mkdir(tmpdir, "subdir")
        writefile(subdir, "file1", "Hello, world!")                            
                                                                                
        if !update(t, fixture, subdir, false) {                                
                t.Log("tally did not report collection changed")                
                t.Fail()                                                        
        }                                                                       
                                                                                
        var collFile = filepath.Join(colldir, "subdir.rscollection")   
        var coll = loadCollection(t, collFile)                                  
        assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
        assertCollectionSize(t, 1, coll)
}

func Test_UpdateSingleDirectory(t *testing.T) {
	fixture := createFixture()
	var config = fixture.GetConfig()
	config.IgnoreWarnings = true
	fixture.SetConfig(config)
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

func Test_UpdateSingleDirectory_addChildren(t *testing.T) {
	fixture := createFixture()
	var config = fixture.GetConfig()
	config.IgnoreWarnings = true
	fixture.SetConfig(config)
	tmpdir := mktmp("Test_UpdateSingleDirectory")
	defer os.RemoveAll(tmpdir)
	
	subdir1 := mkdir(tmpdir, "subdir1")
	writefile(subdir1, "file1", "Hello, world!")

	subdir2 := mkdir(subdir1, "subdir2")
	writefile(subdir2, "file2", "Hello 2")

	subdir3 := mkdir(subdir2, "subdir3")
	writefile(subdir3, "file3", "Hello 3")

	var changed, err = fixture.UpdateSingleDirectory(subdir1,  true)
	if err != nil {
		t.Log("UpdateSingleDirectory failed", err)
		t.Fail()
	}
	if !changed {
		t.Log("Tally did not report a change")
		t.Fail()
	}
	var coll = loadCollectionForDirectory(t, subdir1)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
	assertFileInCollection(t, coll, "subdir2/file2", "465b0f33e43df18353e0395b3c455cf4473f198b")
	assertFileInCollection(t, coll, "subdir2/subdir3/file3", "25f1ffc108e20daaa36d26c3d3d53749d80a6cdf")
	assertCollectionSize(t, 3, coll)
}

func Test_UpdateRecursive_will_fail_when_no_directory(t *testing.T) {
	fixture := createFixture()
	
	var _, err = fixture.UpdateRecursive("this-directory-does-notexist", -1)
	if err == nil {
		t.Log("Should fail when directory does not exist")
		t.Fail()
	}
}

func Test_UpdateRecursive_will_fail_when_no_access(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateRecursive_will_fail_when_no_access")
	defer os.RemoveAll(tmpdir)

	subdir := filepath.Join(tmpdir, "forbidden")
	os.Mkdir(subdir, 0)
	
	var _, err = fixture.UpdateRecursive(subdir, -1)
	if err == nil {
		t.Log("Should fail when directory has incorrect permssions")
		t.Fail()
	}

}

func Test_UpdateRecursive_will_fail_when_no_access_to_subdir(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateRecursive_will_fail_when_no_access_to_subdir")
	defer os.RemoveAll(tmpdir)

	subdir1 := filepath.Join(tmpdir, "subdir1")
	subdir2 := filepath.Join(tmpdir, "subdir1")
	os.Mkdir(subdir2, 0)
	
	var _, err = fixture.UpdateRecursive(subdir1, -1)
	if err == nil {
		t.Log("Should fail when subdirectory has incorrect permssions")
		t.Fail()
	}

}


func Test_UpdateRecursive_will_fail_when_pointed_to_file(t *testing.T) {
	fixture := createFixture()
	tmpdir := mktmp("Test_UpdateSingleDirectory_will_fail_when_pointed_to_file")
	defer os.RemoveAll(tmpdir)

	subdir := mkdir(tmpdir, "forbidden")
	var file =  writefile(subdir, "file1", "Hello, world!")
	
	var _, err = fixture.UpdateRecursive(file, -1)
	if err == nil {
		t.Log("Should fail when target directory is a file")
		t.Fail()
	}
}

func Test_UpdateRecursive_will_fail_when_no_access_to_file(t *testing.T) {
	var fixture = createFixture()

	var tmpdir = mktmp("Test_UpdateRecursive_will_fail_when_no_access_to_file")
	defer os.RemoveAll(tmpdir)

	var subdir = mkdir(tmpdir, "subdir")
	var file =  writefile(subdir, "file1", "Hello, world!")
	os.Chmod(file, 0)
	
	var _, err = fixture.UpdateRecursive(subdir, -1)
	if err == nil {
		t.Log("Should fail when input file has no read permssions")
		t.Fail()
	}
}

func Test_UpdateRecursive_will_fail_when_no_access_to_subfile(t *testing.T) {
	var fixture = createFixture()

	var tmpdir = mktmp("Test_UpdateRecursive_will_fail_when_no_access_to_subfile")
	defer os.RemoveAll(tmpdir)

	var subdir1 = mkdir(tmpdir, "subdir1")
	var subdir2 = mkdir(subdir1, "subdir2")
	var file2 =  writefile(subdir2, "file2", "Hello, world!")
	os.Chmod(file2, 0)
	
	var _, err = fixture.UpdateRecursive(subdir1, -1)
	if err == nil {
		t.Log("Should fail when input file in subdir has no read permssions")
		t.Fail()
	}
}

func Test_UpdateRecursive(t *testing.T) {
	fixture := createFixture()
	var config = fixture.GetConfig()
	config.IgnoreWarnings = true
	config.UpdateParents = true
	fixture.SetConfig(config)
	tmpdir := mktmp("Test_UpdateRecursive")
	defer os.RemoveAll(tmpdir)
	
	subdir1 := mkdir(tmpdir, "subdir1")
	writefile(subdir1, "file1", "Hello, world!")
	var coll = assertUpdateRecursive(t, fixture, subdir1)
	assertCollectionSize(t, 1, coll)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")

	subdir2 := mkdir(subdir1, "subdir2")
	writefile(subdir2, "file2", "Hello 2!")
	coll = assertUpdateRecursive(t, fixture, subdir1)
	assertCollectionSize(t, 2, coll)
	assertFileInCollection(t, coll, "file1", "943a702d06f34599aee1f8da8ef9f7296031d699")
	if coll.ByName("subdir2.rscollection") == nil {
		t.Log("subdir2.rscollection entry not found")
		t.Fail()
	}
	assertWillNotUpdateRecursive(t, fixture, subdir2)
	assertWillNotUpdateRecursive(t, fixture, subdir1)

	var coll2File = resolveCollectionFileSimple(subdir2)
	var sha2 = readShaFor(coll2File)

	subdir3 := mkdir(subdir2, "subdir3")
	assertWillNotUpdateRecursive(t, fixture, subdir2)
	assertWillNotUpdateRecursive(t, fixture, subdir1)
	writefile(subdir3, "file3", "Hello 3!")

	assertUpdateRecursive(t, fixture, subdir3)
	assertShaChanged(t, coll2File, sha2)
}

func resolveCollectionFileSimple(directory string) string {
	var normalized = filepath.Clean(directory)
	return normalized+".rscollection"
}

func Test_UpdateRecursive_will_update_hirerarchy(t *testing.T) {
	var tmpdir = mktmp("Test_UpdateRecursive_will_update_hirerarchy")
	defer os.RemoveAll(tmpdir)
	var fixture = setup_9_directories_testcase(t, tmpdir)
	
	var coll2File = resolveCollectionFileSimple(filepath.Join(tmpdir, "1", "2"))
	var sha2 = readShaFor(coll2File)

	writefile(filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7", "8", "9"), "file9", "changed")
	assertUpdateRecursive(t, fixture, filepath.Join(tmpdir, "1"))

	assertShaChanged(t, coll2File, sha2)
}

func Test_UpdateRecursive_digDepth_0(t *testing.T) {
	var tmpdir = mktmp("Test_UpdateRecursive_will_update_hirerarchy")
	defer os.RemoveAll(tmpdir)
	var fixture = setup_9_directories(t, tmpdir)

	var changed, err = fixture.UpdateRecursive(filepath.Join(tmpdir, "1"), 0)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !changed {
		t.Log("tally did not report a change")
		t.Fail()
	}

	var coll = loadCollection(t, resolveCollectionFileSimple(filepath.Join(tmpdir, "1")))
	assertFileInCollection(t, coll, "file1", "d22ecc269ddb778c3996c096981f0d38fe1a34a9")
	assertFileInCollection(t, coll, "2/file2", "34d7e3c97557e006b730bb979b14fa78ef49d4ed")
	assertFileInCollection(t, coll, "2/3/file3", "ec4767b4b8329f7367256499982d294082554d3d")
	assertFileInCollection(t, coll, "2/3/4/file4", "9b5d4ea26f122f51e321e116fae2dc932fe7d0b0")
	assertFileInCollection(t, coll, "2/3/4/5/file5", "6d66ea6e1f1e2bb2e420b23f302931dc8438d093")
	assertFileInCollection(t, coll, "2/3/4/5/6/file6", "09f736137f24042c16be739d0649051f5ff6950b")
	assertFileInCollection(t, coll, "2/3/4/5/6/7/file7", "3b807d4eeb0fee5cda9aa69092c9dcbc04cf7afe")
	assertFileInCollection(t, coll, "2/3/4/5/6/7/8/file8", "39209371c05e83607c055009d32516b70fcb822c")
	assertFileInCollection(t, coll, "2/3/4/5/6/7/8/9/file9", "41f4bde54a17b753ef14572e7076cb9fb19af2d9")

	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7", "8.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7", "8", "9.rscollection"))
}

func Test_UpdateRecursive_digDepth_1(t *testing.T) {
	var tmpdir = mktmp("Test_UpdateRecursive_will_update_hirerarchy")
	defer os.RemoveAll(tmpdir)
	var fixture = setup_9_directories(t, tmpdir)
	var subdir1 = filepath.Join(tmpdir, "1")

	var changed, err = fixture.UpdateRecursive(subdir1, 1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !changed {
		t.Log("tally did not report a change")
		t.Fail()
	}

	var coll1 = loadCollection(t, resolveCollectionFileSimple(subdir1))
	assertFileInCollection(t, coll1, "2.rscollection", "")

	var subdir2 = filepath.Join(subdir1, "2")
	var coll2 = loadCollection(t, resolveCollectionFileSimple(subdir2)) 

	assertFileInCollection(t, coll2, "3/file3", "ec4767b4b8329f7367256499982d294082554d3d")
	assertFileInCollection(t, coll2, "3/4/file4", "9b5d4ea26f122f51e321e116fae2dc932fe7d0b0")
	assertFileInCollection(t, coll2, "3/4/5/file5", "6d66ea6e1f1e2bb2e420b23f302931dc8438d093")
	assertFileInCollection(t, coll2, "3/4/5/6/file6", "09f736137f24042c16be739d0649051f5ff6950b")
	assertFileInCollection(t, coll2, "3/4/5/6/7/file7", "3b807d4eeb0fee5cda9aa69092c9dcbc04cf7afe")
	assertFileInCollection(t, coll2, "3/4/5/6/7/8/file8", "39209371c05e83607c055009d32516b70fcb822c")
	assertFileInCollection(t, coll2, "3/4/5/6/7/8/9/file9", "41f4bde54a17b753ef14572e7076cb9fb19af2d9")

	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7", "8.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7", "8", "9.rscollection"))
}

func Test_UpdateRecursive_digDepth_2(t *testing.T) {
	var tmpdir = mktmp("Test_UpdateRecursive_will_update_hirerarchy")
	defer os.RemoveAll(tmpdir)
	var fixture = setup_9_directories(t, tmpdir)
	var subdir1 = filepath.Join(tmpdir, "1")

	var changed, err = fixture.UpdateRecursive(subdir1, 2)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !changed {
		t.Log("tally did not report a change")
		t.Fail()
	}

	var coll1 = loadCollection(t, resolveCollectionFileSimple(subdir1))
	assertFileInCollection(t, coll1, "2.rscollection", "")

	var subdir2 = filepath.Join(subdir1, "2")
	var coll2 = loadCollection(t, resolveCollectionFileSimple(subdir2)) 
	assertFileInCollection(t, coll2, "3.rscollection", "")

	var subdir3 = filepath.Join(subdir2, "3")
	var coll3 = loadCollection(t, resolveCollectionFileSimple(subdir3)) 

	assertFileInCollection(t, coll3, "file3", "ec4767b4b8329f7367256499982d294082554d3d")
	assertFileInCollection(t, coll3, "4/file4", "9b5d4ea26f122f51e321e116fae2dc932fe7d0b0")
	assertFileInCollection(t, coll3, "4/5/file5", "6d66ea6e1f1e2bb2e420b23f302931dc8438d093")
	assertFileInCollection(t, coll3, "4/5/6/file6", "09f736137f24042c16be739d0649051f5ff6950b")
	assertFileInCollection(t, coll3, "4/5/6/7/file7", "3b807d4eeb0fee5cda9aa69092c9dcbc04cf7afe")
	assertFileInCollection(t, coll3, "4/5/6/7/8/file8", "39209371c05e83607c055009d32516b70fcb822c")
	assertFileInCollection(t, coll3, "4/5/6/7/8/9/file9", "41f4bde54a17b753ef14572e7076cb9fb19af2d9")

	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7", "8.rscollection"))
	assertPathNotExists(t, filepath.Join(tmpdir, "1", "2", "3", "4", "5", "6", "7", "8", "9.rscollection"))
}

func  Test_UpdateRecursive_will_stop_updating_parents_when_encounters_directory_in_place_of_collection_file(t *testing.T) {
	var tmpdir = mktmp("Test_UpdateRecursive_will_stop_updating_parents_when_encounters_directory_in_place_of_collection_file")
	defer os.RemoveAll(tmpdir)
	var fixture = setup_9_directories_testcase(t, tmpdir)
	
	var coll2File = resolveCollectionFileSimple(filepath.Join(tmpdir, "1", "2"))
	var sha2 = readShaFor(coll2File)

	var coll4File = resolveCollectionFileSimple(filepath.Join(tmpdir, "1", "2", "3", "4"))
	os.RemoveAll(coll4File)
	var err = os.Mkdir(coll4File, os.ModeDir|os.ModePerm)
	if err != nil {
		panic(err)
	}

	var subdir5 = filepath.Join(tmpdir, "1", "2", "3", "4", "5")
	writefile(subdir5, "file5", "change")

	assertUpdateRecursive(t, fixture, subdir5)
	assertShaSame(t, coll2File, sha2)
}

func  Test_UpdateRecursive_will_stop_updating_parents_when_encounters_unreadable_rscollection_in_parent_dir(t *testing.T) {
	var tmpdir = mktmp("Test_UpdateRecursive_will_stop_updating_parents_when_encounters_unreadable_rscollection_in_parent_dir")
	defer os.RemoveAll(tmpdir)
	var fixture = setup_9_directories_testcase(t, tmpdir)

	var coll2File = resolveCollectionFileSimple(filepath.Join(tmpdir, "1", "2"))
	var sha2 = readShaFor(coll2File)

	var coll4File = resolveCollectionFileSimple(filepath.Join(tmpdir, "1", "2", "3", "4"))
	os.Chmod(coll4File, 0)

	var subdir5 = filepath.Join(tmpdir, "1", "2", "3", "4", "5")
	writefile(subdir5, "file5", "change")

	var _,  err = fixture.UpdateRecursive(subdir5, -1)
	if err == nil {
		t.Log("Update should fail due to unreadable rscollection")
		t.Fail()
	}

	assertShaSame(t, coll2File, sha2)
}


func setup_9_directories(t *testing.T, tmpdir string) Tally {
	fixture := createFixture()
	var config = fixture.GetConfig()
	config.IgnoreWarnings = true
	config.UpdateParents = true
	fixture.SetConfig(config)

	subdir1 := mkdir(tmpdir, "1")
	writefile(subdir1, "file1", "Hello 1!")
	subdir2 := mkdir(subdir1, "2")
	writefile(subdir2, "file2", "Hello 2!")
	subdir3 := mkdir(subdir2, "3")
	writefile(subdir3, "file3", "Hello 3!")
	subdir4 := mkdir(subdir3, "4")
	writefile(subdir4, "file4", "Hello 4!")
	subdir5 := mkdir(subdir4, "5")
	writefile(subdir5, "file5", "Hello 5!")
	subdir6 := mkdir(subdir5, "6")
	writefile(subdir6, "file6", "Hello 6!")
	subdir7 := mkdir(subdir6, "7")
	writefile(subdir7, "file7", "Hello 7!")
	subdir8 := mkdir(subdir7, "8")
	writefile(subdir8, "file8", "Hello 8!")
	subdir9 := mkdir(subdir8, "9")
	writefile(subdir9, "file9", "Hello 9!")

	return fixture
}

func setup_9_directories_testcase(t *testing.T, tmpdir string) Tally {
	var fixture = setup_9_directories(t, tmpdir)
	var subdir1 = filepath.Join(tmpdir, "1")
	assertUpdateRecursive(t, fixture, subdir1)
	assertWillNotUpdateRecursive(t, fixture, subdir1)

	return fixture
}

func assertShaSame(t *testing.T, fullpath, oldSha string) {
	var newSha = readShaFor(fullpath)
	if newSha != oldSha {
		t.Log("sha1 for file expeced:", oldSha, "actual:", newSha)
		t.Fail()
	}
}

func assertShaChanged(t *testing.T, fullpath, oldSha string) {
	var newSha = readShaFor(fullpath)
	if newSha == oldSha {
		t.Log("sha1 for file", fullpath, "has not changed")
		t.Fail()
	}
}

func readShaFor(fullpath string) string {
	var dir = filepath.Dir(fullpath)
	var collectionFile =  resolveCollectionFileSimple(dir)
	var fdesc, err = os.Open(collectionFile)
	if err != nil {
		panic(err)
	}
	var coll = NewCollection()
	err = coll.LoadFrom(fdesc)
	var closeErr = fdesc.Close()
	if err != nil {
		panic(err)
	}
	if closeErr != nil {
		panic(closeErr)
	}
	return coll.ByName(filepath.Base(fullpath)).Sha1()
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
	} else if sha1 != "" && sha1 != file.Sha1() {
		t.Log(name, "sha1 is", file.Sha1())
		t.Fail()
	}
}

func assertUpdateRecursive(t *testing.T, fixture Tally, directory string) RSCollection {
	if !update(t, fixture, directory, true) {
		t.Log("tally did not report collection changed")
		t.Fail()
	}
	return loadCollectionForDirectory(t, directory)
}

func assertWillNotUpdateRecursive(t *testing.T, fixture Tally, directory string) {
	var collectionFile = resolveCollectionFileSimple(directory)
	var oldTimestamp = getTimestampSafe(collectionFile)
	if update(t, fixture, directory, true) {
		t.Log("tally reported collection changed while it was not supposed to")
		t.Fail()
	}
	var newTimestamp = getTimestampSafe(collectionFile)
	if oldTimestamp != newTimestamp {
		t.Log("tally should NOT touch collection file!")
	}
}


func assertWillNotUpdateSingleDirectory(t *testing.T, fixture Tally, directory string) {
	var collectionFile = resolveCollectionFileSimple(directory)
	var oldTimestamp = getTimestampSafe(collectionFile)
	if update(t, fixture, directory, false) {
		t.Log("tally reported collection changed while it was not supposed to")
		t.Fail()
	}
	var newTimestamp = getTimestampSafe(collectionFile)
	if oldTimestamp != newTimestamp {
		t.Log("tally should NOT touch collection file!")
	}
}


func assertUpdateSingleDirectory(t *testing.T, fixture Tally, directory string) RSCollection {
	if !update(t, fixture, directory, false) {
		t.Log("tally did not report collection changed")
		t.Fail()
	}
	return loadCollectionForDirectory(t, directory)
}

func update(t *testing.T, fixture Tally, directory string, recursive bool) bool {
	var changed bool
	var err error
	if recursive {
		changed, err = fixture.UpdateRecursive(directory, -1)
	} else {
		changed, err = fixture.UpdateSingleDirectory(directory, false)
	}
	if err != nil {
		t.Log("Cannot update", err)
		t.Fail()
	}
	return changed
}

func loadCollectionForDirectory(t *testing.T, directory string) RSCollection {
	collectionFile := resolveCollectionFileSimple(directory)
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

func Test_resolveRscollectionName(t *testing.T) {
	assertPathNameEvaluatesTo(t, "directory.rscollection", "/path/to/directory", "{{.Path 0}}.rscollection")
	assertPathNameEvaluatesTo(t, "to-directory.rscollection", "/path/to/directory", "{{.Path -1}}-{{.Path 0}}.rscollection")
	assertPathNameEvaluatesTo(t, "-path-to-directory-.rscollection", "/path/to/directory", 
		"{{.Path -3}}-{{.Path -2}}-{{.Path -1}}-{{.Path 0}}-{{.Path 1}}.rscollection")
}

func Test_compileTemplate_InvalidTemplate(t *testing.T) {
	var tallyImpl *tally = createFixture().(*tally)
	var _, err = tallyImpl.compileTemplate("{--{invalid")
	if err != nil {
		t.Log("Should fail on invalid template")
		t.Fail()
	}
}

func Test_executeTemplate_InvalidTemplate(t *testing.T) {
	var tallyImpl *tally = createFixture().(*tally)
	var tpl, err = tallyImpl.compileTemplate("{{.InvalidFunc invalidarg}}")
	if err == nil {
		t.Log("Should not fail on valid template")
		t.Fail()
	}
	var context = mockEvaluationContext("/some/path")
	expanded, err := tallyImpl.executeTemplate(tpl, context)
	if err == nil {
		t.Log("Should not fail on syntactially correct template that could not be executed with proper context")
		t.Fail()
	}
	if expanded != "" {
		t.Log("Should produce empty output")
		t.Fail()
	}
}


func assertPathNameEvaluatesTo(t *testing.T, expected, path, expr string) {
	var tallyImpl = createFixture().(*tally)
	var actual, err = evaluatePathName(tallyImpl, path, expr)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assertStringEquals(t, expected, actual)
}

func assertPathNotExists(t *testing.T, path string) {
	var _, err = os.Stat(path)
	if !os.IsNotExist(err) {
		t.Log("Path", path, "exists")
		t.Fail()
	}
}

func evaluatePathName(tally *tally, path, expr string) (string, error) {
	var context = mockEvaluationContext(path)
	var tpl, err = tally.compileTemplate(expr)
	if err != nil {
		return "", err
	}
	if tpl == nil {
		panic("Returned template is nil")
	}
	var ret string
	ret, err = tally.executeTemplate(tpl, context)
	return ret, err
}

func mockEvaluationContext (path string) TallyPathNameEvalutationContext {
	var ret  = new(pathnameEvaluationContext)
	ret.path = strings.Split(path, string(filepath.Separator))
	return ret
}
