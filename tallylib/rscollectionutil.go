package tallylib

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

func UpdateFile(coll RSCollection, filename string, path string, force bool) (bool, error) {
	var existing RSCollectionFile = nil
	if !force {
		existing = coll.ByPath(filename)
	}

	var stat, err = os.Stat(path)
	if err != nil {
		return false, err
	}

	if shouldUpdate(stat, existing) {
		var digest = sha1.New()

		var file *os.File
		file, err = os.Open(path)
		defer file.Close()

		if _, err = io.Copy(digest, file); err != nil {
			return false, err
		}

		var sha1sum = hex.EncodeToString(digest.Sum(nil))
		if existing == nil || existing.Sha1() != sha1sum {
			coll.Update(filename, sha1sum, stat.Size(),
				stat.ModTime())

			return true, nil
		}
	}

	return false, nil
}

func shouldUpdate(stat os.FileInfo, existing RSCollectionFile) bool {
	return existing == nil || stat.Size() != existing.Size() || stat.ModTime() != existing.Timestamp()
}
