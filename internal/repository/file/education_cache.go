package file

import (
	"context"
	"sync"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

type datasetLoadAttempt struct {
	done chan struct{}
	err  error
}

type lazyDatasetCache[T any] struct {
	mu      sync.Mutex
	loaded  bool
	value   []T
	attempt *datasetLoadAttempt
}

func (c *lazyDatasetCache[T]) get(ctx context.Context, load func(context.Context) ([]T, error), clone func([]T) []T) ([]T, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		c.mu.Lock()
		if c.loaded {
			value := clone(c.value)
			c.mu.Unlock()
			return value, nil
		}
		if c.attempt != nil {
			attempt := c.attempt
			c.mu.Unlock()
			select {
			case <-attempt.done:
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				if attempt.err != nil {
					return nil, attempt.err
				}
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		attempt := &datasetLoadAttempt{done: make(chan struct{})}
		c.attempt = attempt
		c.mu.Unlock()

		value, err := load(ctx)

		c.mu.Lock()
		if err == nil {
			c.value = value
			c.loaded = true
		}
		c.attempt = nil
		attempt.err = err
		close(attempt.done)
		c.mu.Unlock()

		if err != nil {
			return nil, err
		}
		return clone(value), nil
	}
}

func cloneSlice[T any](items []T) []T {
	if len(items) == 0 {
		return make([]T, 0)
	}
	cloned := make([]T, len(items))
	copy(cloned, items)
	return cloned
}

func clonePrimaryAndSecondarySchoolList(schools []models.PrimaryAndSecondarySchool) []models.PrimaryAndSecondarySchool {
	if len(schools) == 0 {
		return make([]models.PrimaryAndSecondarySchool, 0)
	}
	cloned := make([]models.PrimaryAndSecondarySchool, len(schools))
	for i, school := range schools {
		cloned[i] = clonePrimaryAndSecondarySchool(school)
	}
	return cloned
}

func clonePrimaryAndSecondarySchool(school models.PrimaryAndSecondarySchool) models.PrimaryAndSecondarySchool {
	if len(school.EducationLevels) > 0 {
		school.EducationLevels = append([]string(nil), school.EducationLevels...)
	}
	return school
}
