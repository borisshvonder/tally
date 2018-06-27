package tallylib

type Tally interface {
	UpdateSingleDirectory(directory string) error

	// Update all subdirectories and all parent directories (if any)
	UpdateRecursive(directory string) error
}
