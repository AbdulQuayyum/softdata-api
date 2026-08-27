package file

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestJSONRepositoryDecodeObjectsAndArrays(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "object.json"), []byte(`{"name":"kwara","count":4}`))
	writeTestFile(t, filepath.Join(root, "array.json"), []byte(`[1,2,3]`))

	repo, err := NewJSONRepository(root, 64)
	if err != nil {
		t.Fatalf("NewJSONRepository() error = %v", err)
	}

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	var object payload
	if err := repo.Decode(context.Background(), "object.json", &object); err != nil {
		t.Fatalf("Decode(object) error = %v", err)
	}
	if object.Name != "kwara" || object.Count != 4 {
		t.Fatalf("unexpected object: %#v", object)
	}

	var array []int
	if err := repo.Decode(context.Background(), "array.json", &array); err != nil {
		t.Fatalf("Decode(array) error = %v", err)
	}
	if len(array) != 3 || array[0] != 1 || array[2] != 3 {
		t.Fatalf("unexpected array: %#v", array)
	}
}

func TestJSONRepositoryDecodeRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "empty.json"), []byte{})
	writeTestFile(t, filepath.Join(root, "bad.json"), []byte(`{"name":`))
	writeTestFile(t, filepath.Join(root, "trailing.json"), []byte(`{"name":"kwara"} true`))
	writeTestFile(t, filepath.Join(root, "exact.json"), []byte(`[]`))
	writeTestFile(t, filepath.Join(root, "oversize.json"), []byte(`[]x`))

	repo, err := NewJSONRepository(root, 64)
	if err != nil {
		t.Fatalf("NewJSONRepository() error = %v", err)
	}

	if err := repo.Decode(context.Background(), "empty.json", &map[string]any{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("empty file error = %v, want ErrInvalidDatasetFile", err)
	}
	if err := repo.Decode(context.Background(), "bad.json", &map[string]any{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("bad json error = %v, want ErrInvalidDatasetFile", err)
	}
	if err := repo.Decode(context.Background(), "trailing.json", &map[string]any{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("trailing json error = %v, want ErrInvalidDatasetFile", err)
	}

	var value []string
	if err := repo.Decode(context.Background(), "exact.json", &value); err != nil {
		t.Fatalf("Decode(exact) error = %v", err)
	}
	if len(value) != 0 {
		t.Fatalf("unexpected exact value: %#v", value)
	}

	if err := repo.Decode(context.Background(), "exact.json", nil); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("nil destination error = %v, want ErrInvalidDatasetFile", err)
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := repo.Decode(cancelCtx, "object.json", &map[string]any{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled decode error = %v, want context.Canceled", err)
	}

	if err := repo.Decode(context.Background(), "missing.json", &map[string]any{}); !errors.Is(err, interfaces.ErrDatasetFileNotFound) {
		t.Fatalf("missing file error = %v, want ErrDatasetFileNotFound", err)
	}

	for _, err := range []error{repo.Decode(context.Background(), "../escape.json", &map[string]any{}), repo.Decode(context.Background(), string([]byte{0})+".json", &map[string]any{})} {
		if !errors.Is(err, interfaces.ErrInvalidDatasetPath) {
			t.Fatalf("unsafe path error = %v, want ErrInvalidDatasetPath", err)
		}
	}

	sizeRepo, err := NewJSONRepository(root, 2)
	if err != nil {
		t.Fatalf("NewJSONRepository(size) error = %v", err)
	}
	var oversize []string
	if err := sizeRepo.Decode(context.Background(), "oversize.json", &oversize); !errors.Is(err, interfaces.ErrDatasetFileTooLarge) {
		t.Fatalf("oversize error = %v, want ErrDatasetFileTooLarge", err)
	}

	if _, badErr := NewJSONRepository(root, 0); !errors.Is(badErr, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("size constructor error = %v, want ErrInvalidDatasetFile", badErr)
	}

	badErr := repo.Decode(context.Background(), "bad.json", &map[string]any{})
	if strings.Contains(errString(badErr), root) {
		t.Fatal("error leaked root path")
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
