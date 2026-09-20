package app

import (
	"context"
	"errors"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type emergencyAppJSONRepository struct{}

func (emergencyAppJSONRepository) Decode(context.Context, string, any) error { return nil }

type emergencyAppRepository struct{}

func (emergencyAppRepository) ListEmergencyServiceContacts(context.Context, interfaces.EmergencyServiceContactQuery) (interfaces.EmergencyServiceContactListResult, error) {
	return interfaces.EmergencyServiceContactListResult{}, nil
}

func (emergencyAppRepository) GetEmergencyServiceContact(context.Context, string) (models.EmergencyServiceContact, error) {
	return models.EmergencyServiceContact{}, nil
}

func TestEmergencyServiceContactDependencyConstructionUsesDatasetPath(t *testing.T) {
	var gotPath string
	service, err := buildEmergencyServiceContactServiceFromJSONRepository(context.Background(), emergencyAppJSONRepository{},
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.EmergencyServiceContactRepository, error) {
			if repository == nil {
				t.Fatal("json repository was not passed")
			}
			gotPath = recordsPath
			return emergencyAppRepository{}, nil
		},
		func(repository interfaces.EmergencyServiceContactRepository) (emergencyServiceContactService, error) {
			if repository == nil {
				t.Fatal("emergency repository was not passed")
			}
			return services.NewEmergencyServiceContactService(repository)
		},
	)
	if err != nil {
		t.Fatalf("build emergency service contact dependency: %v", err)
	}
	if service == nil {
		t.Fatal("expected service")
	}
	if gotPath != emergencyServiceContactsRelativePath {
		t.Fatalf("unexpected dataset path: got %q want %q", gotPath, emergencyServiceContactsRelativePath)
	}
}

func TestEmergencyServiceContactDependencyConstructionPreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := buildEmergencyServiceContactServiceFromJSONRepository(ctx, emergencyAppJSONRepository{},
		func(interfaces.JSONFileRepository, string) (interfaces.EmergencyServiceContactRepository, error) {
			t.Fatal("repository constructor should not be called")
			return nil, nil
		},
		func(interfaces.EmergencyServiceContactRepository) (emergencyServiceContactService, error) {
			t.Fatal("service constructor should not be called")
			return nil, nil
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
