package file

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var allowedGeoJSONRootTypes = map[string]struct{}{
	"FeatureCollection":  {},
	"Feature":            {},
	"Point":              {},
	"MultiPoint":         {},
	"LineString":         {},
	"MultiLineString":    {},
	"Polygon":            {},
	"MultiPolygon":       {},
	"GeometryCollection": {},
}

// GeoJSONRepository validates and returns GeoJSON documents from files.
type GeoJSONRepository struct {
	store *safeStore
}

// NewGeoJSONRepository constructs a GeoJSON repository rooted at the supplied path.
func NewGeoJSONRepository(root string, maxBytes int64) (*GeoJSONRepository, error) {
	store, err := newSafeStore(root, maxBytes)
	if err != nil {
		return nil, err
	}
	return &GeoJSONRepository{store: store}, nil
}

// Read loads and validates a GeoJSON document without altering its representation.
func (r *GeoJSONRepository) Read(ctx context.Context, relativePath string) (json.RawMessage, error) {
	if r == nil || r.store == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	data, err := r.store.readBytes(ctx, relativePath)
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var root struct {
		Type string `json:"type"`
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&root); err != nil {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := ensureTrailingJSONWhitespaceOnly(dec); err != nil {
		return nil, err
	}

	typeName := strings.TrimSpace(root.Type)
	if typeName == "" {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if _, ok := allowedGeoJSONRootTypes[typeName]; !ok {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	return append(json.RawMessage(nil), data...), nil
}

func ensureTrailingJSONWhitespaceOnly(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}
