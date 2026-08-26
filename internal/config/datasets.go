package config

import (
	"path/filepath"
	"strings"
)

type DatasetsConfig struct {
	Path string
}

func loadDatasetsConfig() (DatasetsConfig, error) {
	path := strings.TrimSpace(getEnv("DATASETS_PATH"))
	if path == "" {
		path = "datasets"
	}

	return DatasetsConfig{
		Path: filepath.Clean(path),
	}, nil
}
