package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type safeStore struct {
	rootReal string
	maxBytes int64
}

func newSafeStore(root string, maxBytes int64) (*safeStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}
	if maxBytes <= 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	absRoot, err := filepath.Abs(strings.TrimSpace(root))
	if err != nil {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}

	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}

	info, err := os.Stat(realRoot)
	if err != nil {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}

	return &safeStore{
		rootReal: realRoot,
		maxBytes: maxBytes,
	}, nil
}

func (s *safeStore) resolve(relativePath string) (string, error) {
	if s == nil {
		return "", fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if relativePath == "" {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}
	if strings.ContainsRune(relativePath, 0) {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}
	if filepath.IsAbs(relativePath) || filepath.VolumeName(relativePath) != "" {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}

	cleaned := filepath.Clean(relativePath)
	if cleaned == "." {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}

	candidate := filepath.Join(s.rootReal, cleaned)
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", classifyPathError(err)
	}

	rel, err := filepath.Rel(s.rootReal, resolved)
	if err != nil {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidDatasetPath)
	}

	return resolved, nil
}

func (s *safeStore) readBytes(ctx context.Context, relativePath string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	resolved, err := s.resolve(relativePath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return nil, classifyPathError(err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if info.Size() > s.maxBytes {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileTooLarge)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := os.Open(resolved)
	if err != nil {
		return nil, classifyPathError(err)
	}
	defer file.Close()

	limit := s.maxBytes + 1
	if limit < 0 {
		limit = s.maxBytes
	}

	data, err := io.ReadAll(io.LimitReader(file, limit))
	if err != nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if int64(len(data)) > s.maxBytes {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileTooLarge)
	}

	return append([]byte(nil), data...), nil
}

func classifyPathError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("%w", interfaces.ErrDatasetFileNotFound)
	case errors.Is(err, fs.ErrPermission):
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	default:
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
}

func rejectNilDestination(destination any) bool {
	if destination == nil {
		return true
	}
	value := reflect.ValueOf(destination)
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func:
		return value.IsNil()
	default:
		return false
	}
}
