package file

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestGeoJSONRepositoryAcceptsStandardRootTypes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	repo, err := NewGeoJSONRepository(root, 128)
	if err != nil {
		t.Fatalf("NewGeoJSONRepository() error = %v", err)
	}

	types := []string{
		"FeatureCollection",
		"Feature",
		"Point",
		"MultiPoint",
		"LineString",
		"MultiLineString",
		"Polygon",
		"MultiPolygon",
		"GeometryCollection",
	}

	for _, typeName := range types {
		typeName := typeName
		t.Run(typeName, func(t *testing.T) {
			t.Parallel()

			fileName := strings.ToLower(typeName) + ".geojson"
			contents := []byte(`{"type":"` + typeName + `"}`)
			writeTestFile(t, filepath.Join(root, fileName), contents)

			raw, err := repo.Read(context.Background(), fileName)
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}
			if !bytes.Equal(raw, contents) {
				t.Fatalf("unexpected raw message: %q", []byte(raw))
			}

			raw[0] = ' '
			again, err := repo.Read(context.Background(), fileName)
			if err != nil {
				t.Fatalf("Read() again error = %v", err)
			}
			if !bytes.Equal(again, contents) {
				t.Fatalf("returned raw bytes shared mutable storage: %q", []byte(again))
			}
		})
	}
}

func TestGeoJSONRepositoryRejectsInvalidDocuments(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "empty.geojson"), []byte{})
	writeTestFile(t, filepath.Join(root, "missing.geojson"), []byte(`{}`))
	writeTestFile(t, filepath.Join(root, "nonstr.geojson"), []byte(`{"type":1}`))
	writeTestFile(t, filepath.Join(root, "unsupported.geojson"), []byte(`{"type":"Topology"}`))
	writeTestFile(t, filepath.Join(root, "trailing.geojson"), []byte(`{"type":"Feature"} {"extra":true}`))

	repo, err := NewGeoJSONRepository(root, 64)
	if err != nil {
		t.Fatalf("NewGeoJSONRepository() error = %v", err)
	}

	for _, name := range []string{"empty.geojson", "missing.geojson", "nonstr.geojson", "unsupported.geojson", "trailing.geojson"} {
		if _, err := repo.Read(context.Background(), name); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
			t.Fatalf("Read(%q) error = %v, want ErrInvalidDatasetFile", name, err)
		}
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.Read(cancelCtx, "unsupported.geojson"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled read error = %v, want context.Canceled", err)
	}

	if _, err := repo.Read(context.Background(), "missing-file.geojson"); !errors.Is(err, interfaces.ErrDatasetFileNotFound) {
		t.Fatalf("missing file error = %v, want ErrDatasetFileNotFound", err)
	}

	if _, err := repo.Read(context.Background(), "../escape.geojson"); !errors.Is(err, interfaces.ErrInvalidDatasetPath) {
		t.Fatalf("unsafe path error = %v, want ErrInvalidDatasetPath", err)
	}

	if _, err := repo.Read(context.Background(), string([]byte{0})+".geojson"); !errors.Is(err, interfaces.ErrInvalidDatasetPath) {
		t.Fatalf("nul path error = %v, want ErrInvalidDatasetPath", err)
	}
}
