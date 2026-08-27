package config

import (
	"path/filepath"
)

type DatasetConfig struct {
	Path string
}

type DatasetsConfig = DatasetConfig

func loadDatasetsConfig(lookup LookupEnv) (DatasetConfig, error) {
	path := lookupString(lookup, "DATASETS_PATH")
	if path == "" {
		path = "datasets"
	}

	return DatasetConfig{
		Path: filepath.Clean(path),
	}, nil
}
