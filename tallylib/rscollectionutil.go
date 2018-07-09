package tallylib

import (
	"encoding/hex"
	"crypto/sha1"
	"io"
	"os"
)

func updateFile(coll RSCollection, name string, path string, force bool) (bool, error) {
	var existing RSCollectionFile = nil
	if !force {
		existing = coll.ByName(name)
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
			coll.Update(name, sha1sum, stat.Size(), stat.ModTime())
			return true, nil
		}
	}

	return false, nil
}

func shouldUpdate(stat os.FileInfo, existing RSCollectionFile) bool {
	return existing == nil || stat.Size() != existing.Size() || stat.ModTime() != existing.Timestamp()
}
