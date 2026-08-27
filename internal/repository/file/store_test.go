package file

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func writeTestFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func isNilValue(value any) bool {
	if value == nil {
		return true
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}

func TestConstructorsValidateRootsAndSize(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	fileRoot := filepath.Join(t.TempDir(), "root-file.txt")
	writeTestFile(t, fileRoot, []byte("not a directory"))
	missingRoot := filepath.Join(t.TempDir(), "missing")

	cases := []struct {
		name string
		ctor func(string, int64) (any, error)
	}{
		{name: "json", ctor: func(root string, maxBytes int64) (any, error) { return NewJSONRepository(root, maxBytes) }},
		{name: "csv", ctor: func(root string, maxBytes int64) (any, error) { return NewCSVRepository(root, maxBytes) }},
		{name: "geojson", ctor: func(root string, maxBytes int64) (any, error) { return NewGeoJSONRepository(root, maxBytes) }},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if repo, err := tc.ctor(root, 1); err != nil || isNilValue(repo) {
				t.Fatalf("constructor valid root error = %v repo = %T", err, repo)
			}

			for _, bad := range []struct {
				name string
				root string
			}{
				{name: "empty", root: ""},
				{name: "missing", root: missingRoot},
				{name: "file", root: fileRoot},
			} {
				t.Run(bad.name, func(t *testing.T) {
					t.Parallel()

					repo, err := tc.ctor(bad.root, 1)
					if err == nil {
						t.Fatalf("constructor error = nil, want error")
					}
					if !isNilValue(repo) {
						t.Fatalf("constructor repo = %#v, want nil", repo)
					}
					if (bad.root != "" && strings.Contains(err.Error(), bad.root)) || strings.Contains(err.Error(), root) {
						t.Fatalf("constructor leaked root path: %v", err)
					}
					if !errors.Is(err, interfaces.ErrInvalidDatasetPath) {
						t.Fatalf("constructor error = %v, want ErrInvalidDatasetPath", err)
					}
				})
			}

			repo, err := tc.ctor(root, 0)
			if err == nil {
				t.Fatal("constructor error = nil, want size validation error")
			}
			if !isNilValue(repo) {
				t.Fatalf("constructor repo = %#v, want nil", repo)
			}
			if !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("constructor error = %v, want ErrInvalidDatasetFile", err)
			}
		})
	}
}

func TestSafeStorePathResolutionAndSizeChecks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatalf("Mkdir(%q) error = %v", nested, err)
	}

	exactPath := filepath.Join(nested, "exact.txt")
	writeTestFile(t, exactPath, []byte("abcd"))

	tooLargePath := filepath.Join(nested, "too-large.txt")
	writeTestFile(t, tooLargePath, []byte("abcde"))

	directoryPath := filepath.Join(nested, "dir")
	if err := os.Mkdir(directoryPath, 0o755); err != nil {
		t.Fatalf("Mkdir(%q) error = %v", directoryPath, err)
	}

	outsideRoot := filepath.Join(t.TempDir(), "outside.txt")
	writeTestFile(t, outsideRoot, []byte("outside"))

	symlinkPath := filepath.Join(nested, "escape.txt")
	if err := os.Symlink(outsideRoot, symlinkPath); err != nil {
		t.Skipf("Symlink not available: %v", err)
	}

	store, err := newSafeStore(root, 4)
	if err != nil {
		t.Fatalf("newSafeStore() error = %v", err)
	}

	data, err := store.readBytes(context.Background(), filepath.ToSlash(filepath.Join("nested", "exact.txt")))
	if err != nil {
		t.Fatalf("readBytes() error = %v", err)
	}
	if string(data) != "abcd" {
		t.Fatalf("unexpected data: %q", string(data))
	}

	oversized, err := store.readBytes(context.Background(), filepath.ToSlash(filepath.Join("nested", "too-large.txt")))
	if err == nil {
		t.Fatal("readBytes() error = nil, want oversized error")
	}
	if oversized != nil {
		t.Fatalf("readBytes() data = %q, want nil", string(oversized))
	}
	if !errors.Is(err, interfaces.ErrDatasetFileTooLarge) {
		t.Fatalf("readBytes() error = %v, want ErrDatasetFileTooLarge", err)
	}

	if _, err := store.readBytes(context.Background(), filepath.ToSlash(filepath.Join("nested", "dir"))); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("directory read error = %v, want ErrInvalidDatasetFile", err)
	}

	if _, err := store.readBytes(context.Background(), "missing.txt"); !errors.Is(err, interfaces.ErrDatasetFileNotFound) {
		t.Fatalf("missing file error = %v, want ErrDatasetFileNotFound", err)
	}

	for _, rel := range []string{
		"../outside.txt",
		filepath.ToSlash(filepath.Join("..", "outside.txt")),
		string(filepath.Separator) + "absolute.txt",
		"nested/../..",
		"bad\x00path",
		"nested/escape.txt",
	} {
		_, err := store.readBytes(context.Background(), rel)
		if err == nil {
			t.Fatalf("readBytes(%q) error = nil, want error", rel)
		}
		if strings.Contains(err.Error(), root) || strings.Contains(err.Error(), outsideRoot) {
			t.Fatalf("readBytes(%q) leaked filesystem path: %v", rel, err)
		}
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.readBytes(cancelCtx, filepath.ToSlash(filepath.Join("nested", "exact.txt"))); !errors.Is(err, context.Canceled) {
		t.Fatalf("readBytes() error = %v, want context.Canceled", err)
	}

	if _, err := store.readBytes(context.Background(), filepath.ToSlash(filepath.Join("nested", "exact.txt"))); err != nil {
		t.Fatalf("second readBytes() error = %v", err)
	}

	resolved, err := store.resolve(filepath.ToSlash(filepath.Join("nested", "exact.txt")))
	if err != nil {
		t.Fatalf("resolve() error = %v", err)
	}
	if !filepath.IsAbs(resolved) {
		t.Fatalf("resolve() path = %q, want absolute path", resolved)
	}

	_, err = store.resolve("nested/escape.txt")
	if !errors.Is(err, interfaces.ErrInvalidDatasetPath) {
		t.Fatalf("resolve() error = %v, want ErrInvalidDatasetPath", err)
	}
	if !errors.Is(classifyPathError(fs.ErrNotExist), interfaces.ErrDatasetFileNotFound) {
		t.Fatalf("classifyPathError() sentinel mismatch")
	}
}
