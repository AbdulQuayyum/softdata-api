package file

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestCSVRepositoryReadWithAndWithoutHeader(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "states.csv"), []byte("name,notes\nKwara,\"line1\nline2\"\nLagos,\"a,b\"\n"))

	repo, err := NewCSVRepository(root, 128)
	if err != nil {
		t.Fatalf("NewCSVRepository() error = %v", err)
	}

	withHeader, err := repo.Read(context.Background(), "states.csv", interfaces.CSVReadOptions{HasHeader: true})
	if err != nil {
		t.Fatalf("Read(has header) error = %v", err)
	}
	if !reflect.DeepEqual(withHeader.Header, []string{"name", "notes"}) {
		t.Fatalf("unexpected header: %#v", withHeader.Header)
	}
	if len(withHeader.Records) != 2 {
		t.Fatalf("unexpected record count: %d", len(withHeader.Records))
	}
	if withHeader.Records[0][0] != "Kwara" || withHeader.Records[0][1] != "line1\nline2" {
		t.Fatalf("unexpected first record: %#v", withHeader.Records[0])
	}
	if withHeader.Records[1][0] != "Lagos" || withHeader.Records[1][1] != "a,b" {
		t.Fatalf("unexpected second record: %#v", withHeader.Records[1])
	}

	headerless, err := repo.Read(context.Background(), "states.csv", interfaces.CSVReadOptions{HasHeader: false})
	if err != nil {
		t.Fatalf("Read(headerless) error = %v", err)
	}
	if len(headerless.Header) != 0 {
		t.Fatalf("unexpected headerless header: %#v", headerless.Header)
	}
	if len(headerless.Records) != 3 {
		t.Fatalf("unexpected headerless record count: %d", len(headerless.Records))
	}
	if headerless.Records[0][0] != "name" || headerless.Records[0][1] != "notes" {
		t.Fatalf("unexpected headerless first record: %#v", headerless.Records[0])
	}
}

func TestCSVRepositoryRejectsEmptyMalformedAndCopiesRows(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "empty.csv"), []byte{})
	writeTestFile(t, filepath.Join(root, "bad.csv"), []byte("name,\"unterminated"))
	writeTestFile(t, filepath.Join(root, "header-only.csv"), []byte("name,code\n"))

	repo, err := NewCSVRepository(root, 32)
	if err != nil {
		t.Fatalf("NewCSVRepository() error = %v", err)
	}

	if _, err := repo.Read(context.Background(), "empty.csv", interfaces.CSVReadOptions{HasHeader: true}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("empty csv error = %v, want ErrInvalidDatasetFile", err)
	}
	if _, err := repo.Read(context.Background(), "bad.csv", interfaces.CSVReadOptions{HasHeader: true}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("bad csv error = %v, want ErrInvalidDatasetFile", err)
	}

	headerOnly, err := repo.Read(context.Background(), "header-only.csv", interfaces.CSVReadOptions{HasHeader: true})
	if err != nil {
		t.Fatalf("Read(header only) error = %v", err)
	}
	if headerOnly.Records == nil || len(headerOnly.Records) != 0 {
		t.Fatalf("unexpected header-only records: %#v", headerOnly.Records)
	}
	if !reflect.DeepEqual(headerOnly.Header, []string{"name", "code"}) {
		t.Fatalf("unexpected header-only header: %#v", headerOnly.Header)
	}

	rows, err := repo.Read(context.Background(), "header-only.csv", interfaces.CSVReadOptions{HasHeader: false})
	if err != nil {
		t.Fatalf("Read(header only without header) error = %v", err)
	}
	rows.Records[0][0] = "changed"
	rows.Header = []string{"changed"}
	again, err := repo.Read(context.Background(), "header-only.csv", interfaces.CSVReadOptions{HasHeader: false})
	if err != nil {
		t.Fatalf("Read again error = %v", err)
	}
	if again.Records[0][0] != "name" {
		t.Fatalf("returned rows shared mutable storage: %#v", again.Records)
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.Read(cancelCtx, "header-only.csv", interfaces.CSVReadOptions{HasHeader: true}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled read error = %v, want context.Canceled", err)
	}

	if _, err := repo.Read(context.Background(), "missing.csv", interfaces.CSVReadOptions{HasHeader: true}); !errors.Is(err, interfaces.ErrDatasetFileNotFound) {
		t.Fatalf("missing csv error = %v, want ErrDatasetFileNotFound", err)
	}

	unsafeErr := func() error {
		_, err := repo.Read(context.Background(), "../escape.csv", interfaces.CSVReadOptions{HasHeader: true})
		return err
	}()
	if !errors.Is(unsafeErr, interfaces.ErrInvalidDatasetPath) {
		t.Fatalf("unsafe csv path error = %v, want ErrInvalidDatasetPath", unsafeErr)
	}

	if strings.Contains(errString(unsafeErr), root) {
		t.Fatal("csv error leaked root path")
	}
}
