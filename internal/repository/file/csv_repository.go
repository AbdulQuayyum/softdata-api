package file

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

// CSVRepository reads CSV documents from files beneath a dataset root.
type CSVRepository struct {
	store *safeStore
}

// NewCSVRepository constructs a CSV repository rooted at the supplied path.
func NewCSVRepository(root string, maxBytes int64) (*CSVRepository, error) {
	store, err := newSafeStore(root, maxBytes)
	if err != nil {
		return nil, err
	}
	return &CSVRepository{store: store}, nil
}

// Read loads a CSV file using explicit header handling.
func (r *CSVRepository) Read(ctx context.Context, relativePath string, options interfaces.CSVReadOptions) (interfaces.CSVDocument, error) {
	if r == nil || r.store == nil {
		return interfaces.CSVDocument{}, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return interfaces.CSVDocument{}, err
	}

	data, err := r.store.readBytes(ctx, relativePath)
	if err != nil {
		return interfaces.CSVDocument{}, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return interfaces.CSVDocument{}, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := ctx.Err(); err != nil {
		return interfaces.CSVDocument{}, err
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	reader.ReuseRecord = false

	document := interfaces.CSVDocument{Records: make([][]string, 0)}
	if options.HasHeader {
		header, err := reader.Read()
		if err != nil {
			return interfaces.CSVDocument{}, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		document.Header = append([]string(nil), header...)
	}

	for {
		if err := ctx.Err(); err != nil {
			return interfaces.CSVDocument{}, err
		}
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return interfaces.CSVDocument{}, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		document.Records = append(document.Records, append([]string(nil), record...))
	}

	return document, nil
}
