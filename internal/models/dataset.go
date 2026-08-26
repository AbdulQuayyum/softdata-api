package models

import "time"

type DatasetStatus string

const (
	DatasetStatusDraft      DatasetStatus = "draft"
	DatasetStatusReview     DatasetStatus = "review"
	DatasetStatusActive     DatasetStatus = "active"
	DatasetStatusDeprecated DatasetStatus = "deprecated"
	DatasetStatusArchived   DatasetStatus = "archived"
)

// Dataset is the internal persistence model for dataset metadata.
type Dataset struct {
	ID string `json:"-"`
	// DatasetKey is the stable public identifier exposed as id in API responses.
	DatasetKey      string `json:"-"`
	Slug            string
	Name            string
	Description     *string
	GroupName       string
	CountryCode     *string
	Version         string
	Status          DatasetStatus
	RecordCount     int64
	PrimaryFormat   string
	Formats         []string
	SchemaPath      *string
	LicenceID       *string
	SourceCount     int64
	UpdateFrequency *string
	LastUpdatedAt   *time.Time
	LastVerifiedAt  *time.Time
	Maintainers     []string
	IsPublic        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ArchivedAt      *time.Time
}

// DatasetResponse is the public dataset payload where id maps to DatasetKey.
type DatasetResponse struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Description     *string       `json:"description,omitempty"`
	Group           string        `json:"group"`
	CountryCode     *string       `json:"country_code,omitempty"`
	Version         string        `json:"version"`
	Status          DatasetStatus `json:"status"`
	RecordCount     int64         `json:"record_count"`
	PrimaryFormat   string        `json:"primary_format"`
	Formats         []string      `json:"formats"`
	Schema          *string       `json:"schema,omitempty"`
	SourceIDs       []string      `json:"source_ids,omitempty"`
	LicenceID       *string       `json:"licence_id,omitempty"`
	SourceCount     int64         `json:"source_count"`
	UpdateFrequency *string       `json:"update_frequency,omitempty"`
	LastUpdatedAt   *string       `json:"last_updated_at,omitempty"`
	LastVerifiedAt  *string       `json:"last_verified_at,omitempty"`
	Maintainers     []string      `json:"maintainers"`
	IsPublic        bool          `json:"is_public"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	ArchivedAt      *time.Time    `json:"archived_at,omitempty"`
}
