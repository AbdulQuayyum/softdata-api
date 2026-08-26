package models

import "time"

// DatasetSource captures provenance metadata for a dataset.
type DatasetSource struct {
	ID             string `json:"-"`
	DatasetID      string `json:"-"`
	SourceKey      string
	Name           string
	URL            *string
	Description    *string
	Publisher      *string
	SourceType     *string
	LicenceID      *string
	IsOfficial     bool
	LastFetchedAt  *time.Time
	LastVerifiedAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// DatasetSourceResponse is the public source representation.
type DatasetSourceResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	URL            *string    `json:"url,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Publisher      *string    `json:"publisher,omitempty"`
	SourceType     *string    `json:"source_type,omitempty"`
	LicenceID      *string    `json:"licence_id,omitempty"`
	IsOfficial     bool       `json:"is_official"`
	LastFetchedAt  *time.Time `json:"last_fetched_at,omitempty"`
	LastVerifiedAt *time.Time `json:"last_verified_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
