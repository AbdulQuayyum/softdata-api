package interfaces

import (
	"context"
	"encoding/json"
)

// CSVReadOptions controls how CSV files are interpreted.
type CSVReadOptions struct {
	HasHeader bool
}

// CSVDocument is the neutral CSV result returned by file repositories.
type CSVDocument struct {
	Header  []string
	Records [][]string
}

// JSONFileRepository decodes a JSON document from a dataset file.
type JSONFileRepository interface {
	Decode(ctx context.Context, relativePath string, destination any) error
}

// CSVFileRepository reads a CSV document from a dataset file.
type CSVFileRepository interface {
	Read(ctx context.Context, relativePath string, options CSVReadOptions) (CSVDocument, error)
}

// GeoJSONFileRepository reads a validated GeoJSON document from a dataset file.
type GeoJSONFileRepository interface {
	Read(ctx context.Context, relativePath string) (json.RawMessage, error)
}
