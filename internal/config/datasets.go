package config

import (
	"path/filepath"
)

type DatasetConfig struct {
	Path         string
	JSONMaxBytes int64
}

type DatasetsConfig = DatasetConfig

func loadDatasetsConfig(lookup LookupEnv) (DatasetConfig, error) {
	path := lookupString(lookup, "DATASETS_PATH")
	if path == "" {
		path = "datasets"
	}
	jsonMaxBytes, err := parsePositiveInt64("DATASETS_JSON_MAX_BYTES", lookupString(lookup, "DATASETS_JSON_MAX_BYTES"), 16777216)
	if err != nil {
		return DatasetConfig{}, err
	}

	return DatasetConfig{
		Path:         filepath.Clean(path),
		JSONMaxBytes: jsonMaxBytes,
	}, nil
}
