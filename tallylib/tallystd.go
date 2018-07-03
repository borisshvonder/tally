package tallylib

import (
//	"os"
)

// Holds settings
type tally struct {
	config TallyConfig
}

func (tally *tally) GetConfig() TallyConfig {
	return tally.config
}

func (tally *tally) SetConfig(cfg TallyConfig) {
	tally.config = cfg
}

func (tally *tally) UpdateSingleDirectory(directory string) error {
	return nil
}

func (tally *tally) UpdateRecursive(directory string) error {
	return nil
}
