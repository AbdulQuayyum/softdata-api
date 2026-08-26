package models

import "time"

type DatasetVersionStatus string

const (
	DatasetVersionStatusDraft      DatasetVersionStatus = "draft"
	DatasetVersionStatusPublished  DatasetVersionStatus = "published"
	DatasetVersionStatusDeprecated DatasetVersionStatus = "deprecated"
	DatasetVersionStatusArchived   DatasetVersionStatus = "archived"
)

// DatasetVersion captures versioned dataset metadata and release tracking.
type DatasetVersion struct {
	ID            string `json:"-"`
	DatasetID     string `json:"-"`
	Version       string
	SchemaVersion *string
	Format        string
	Status        DatasetVersionStatus
	RecordCount   int64
	Checksum      *string
	StoragePath   *string
	Notes         *string
	ReleasedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// DatasetVersionResponse is the public version representation.
type DatasetVersionResponse struct {
	ID            string               `json:"id"`
	DatasetID     string               `json:"dataset_id"`
	Version       string               `json:"version"`
	SchemaVersion *string              `json:"schema_version,omitempty"`
	Format        string               `json:"format"`
	Status        DatasetVersionStatus `json:"status"`
	RecordCount   int64                `json:"record_count"`
	Checksum      *string              `json:"checksum,omitempty"`
	Notes         *string              `json:"notes,omitempty"`
	ReleasedAt    *string              `json:"released_at,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}
