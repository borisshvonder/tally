package tallylib

type TallyConfig struct {
	forceUpdate bool
}

type Tally interface {
	GetConfig() TallyConfig
	SetConfig(cfg TallyConfig)

	UpdateSingleDirectory(directory string) error

	// Update all subdirectories and all parent directories (if any)
	UpdateRecursive(directory string) error
}
