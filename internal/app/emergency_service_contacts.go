package app

import (
	"context"
	"fmt"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func buildEmergencyServiceContactServiceFromJSONRepository(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newRepository func(interfaces.JSONFileRepository, string) (interfaces.EmergencyServiceContactRepository, error),
	newService func(interfaces.EmergencyServiceContactRepository) (emergencyServiceContactService, error),
) (emergencyServiceContactService, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	if newRepository == nil {
		return nil, fmt.Errorf("emergency service contact repository constructor is required")
	}
	if newService == nil {
		return nil, fmt.Errorf("emergency service contact service constructor is required")
	}
	repository, err := newRepository(jsonRepository, emergencyServiceContactsRelativePath)
	if err != nil {
		return nil, fmt.Errorf("initialize emergency service contact repository: %w", err)
	}
	service, err := newService(repository)
	if err != nil {
		return nil, fmt.Errorf("initialize emergency service contact service: %w", err)
	}
	return service, nil
}
