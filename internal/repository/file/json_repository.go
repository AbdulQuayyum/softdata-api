package file

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

// JSONRepository decodes JSON documents from files beneath a dataset root.
type JSONRepository struct {
	store *safeStore
}

// NewJSONRepository constructs a JSON repository rooted at the supplied path.
func NewJSONRepository(root string, maxBytes int64) (*JSONRepository, error) {
	store, err := newSafeStore(root, maxBytes)
	if err != nil {
		return nil, err
	}
	return &JSONRepository{store: store}, nil
}

// Decode reads and decodes exactly one JSON value into destination.
func (r *JSONRepository) Decode(ctx context.Context, relativePath string, destination any) error {
	if r == nil || r.store == nil {
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if rejectNilDestination(destination) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	data, err := r.store.readBytes(ctx, relativePath)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(destination); err != nil {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}
