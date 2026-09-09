package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type metadata struct {
	DatasetKey      string   `json:"dataset_key"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	CountryCode     string   `json:"country_code"`
	DatasetGroup    string   `json:"dataset_group"`
	Format          string   `json:"format"`
	RelativePath    string   `json:"relative_path"`
	SchemaPath      string   `json:"schema_path"`
	RecordCount     int64    `json:"record_count"`
	Version         string   `json:"version"`
	LicenseID       string   `json:"license_id"`
	UpdateFrequency string   `json:"update_frequency"`
	Maintainers     []string `json:"maintainers"`
	VerifiedAt      string   `json:"verified_at"`
	Status          string   `json:"status"`
	Sources         []source `json:"sources"`
}

type source struct {
	Organization string `json:"organization"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	Purpose      string `json:"purpose"`
	AccessedAt   string `json:"accessed_at"`
}

func main() {
	metadataDir := flag.String("metadata-dir", "datasets/metadata", "directory containing dataset metadata JSON files")
	flag.Parse()

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = loadEnvValue(".env", "DATABASE_URL")
	}
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	items, err := loadMetadata(*metadataDir)
	if err != nil {
		log.Fatalf("load metadata: %v", err)
	}
	if len(items) == 0 {
		log.Fatal("no canonical dataset metadata found")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("begin transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, item := range items {
		if err := upsertDataset(ctx, tx, item); err != nil {
			log.Fatalf("upsert dataset %q: %v", item.DatasetKey, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit dataset catalog: %v", err)
	}

	log.Printf("seeded %d datasets", len(items))
}

func loadEnvValue(path, wantedKey string) string {
	contents, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != wantedKey {
			continue
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"'")
		return value
	}
	return ""
}

func loadMetadata(root string) ([]metadata, error) {
	items := make([]metadata, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var fields map[string]json.RawMessage
		if err := json.Unmarshal(contents, &fields); err != nil {
			return fmt.Errorf("decode %s: %w", path, err)
		}
		if len(fields["dataset_key"]) == 0 || len(fields["relative_path"]) == 0 {
			return nil
		}

		var item metadata
		if err := json.Unmarshal(contents, &item); err != nil {
			return fmt.Errorf("decode %s: %w", path, err)
		}
		if item.DatasetKey == "" || item.RelativePath == "" {
			return nil
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func upsertDataset(ctx context.Context, tx pgx.Tx, item metadata) error {
	version := item.Version
	if version == "" {
		version = "1.0.0"
	}
	status := item.Status
	if status == "" {
		status = "active"
	}
	format := item.Format
	if format == "" {
		format = "json"
	}
	group := item.DatasetGroup
	if group == "" {
		group = "other"
	}
	countryCode := item.CountryCode
	if len(countryCode) != 2 {
		countryCode = ""
	}
	formats := []string{format}
	verifiedAt := nullableDate(item.VerifiedAt)
	lastUpdatedAt := verifiedAt

	var datasetID string
	err := tx.QueryRow(ctx, `
		INSERT INTO datasets (
			dataset_key, slug, name, description, group_name, country_code,
			version, status, record_count, primary_format, formats, schema_path,
			licence_id, source_count, update_frequency, last_updated_at,
			last_verified_at, maintainers, is_public
		) VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), $7, $8,
			$9, $10, $11, NULLIF($12, ''), NULLIF($13, ''), $14,
			NULLIF($15, ''), $16, $16, COALESCE($17, '{}'::text[]), true)
		ON CONFLICT (lower(dataset_key)) DO UPDATE SET
			slug = EXCLUDED.slug,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			group_name = EXCLUDED.group_name,
			country_code = EXCLUDED.country_code,
			version = EXCLUDED.version,
			status = EXCLUDED.status,
			record_count = EXCLUDED.record_count,
			primary_format = EXCLUDED.primary_format,
			formats = EXCLUDED.formats,
			schema_path = EXCLUDED.schema_path,
			licence_id = EXCLUDED.licence_id,
			source_count = EXCLUDED.source_count,
			update_frequency = EXCLUDED.update_frequency,
			last_updated_at = EXCLUDED.last_updated_at,
			last_verified_at = EXCLUDED.last_verified_at,
			maintainers = EXCLUDED.maintainers,
			is_public = EXCLUDED.is_public,
			updated_at = now(),
			archived_at = NULL
		RETURNING id::text`,
		item.DatasetKey, item.DatasetKey, item.Title, item.Description, group,
		countryCode, version, status, item.RecordCount, format, formats,
		item.SchemaPath, item.LicenseID, len(item.Sources), item.UpdateFrequency,
		lastUpdatedAt, item.Maintainers,
	).Scan(&datasetID)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM dataset_sources WHERE dataset_id = $1`, datasetID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM dataset_versions WHERE dataset_id = $1`, datasetID); err != nil {
		return err
	}

	for index, itemSource := range item.Sources {
		accessedAt := nullableDate(itemSource.AccessedAt)
		sourceName := itemSource.Title
		if sourceName == "" {
			sourceName = itemSource.Organization
		}
		if sourceName == "" {
			sourceName = itemSource.URL
		}
		if sourceName == "" {
			sourceName = item.Title
		}
		sourceKey := fmt.Sprintf("%s-%d", slugify(sourceName), index+1)
		if _, err := tx.Exec(ctx, `
			INSERT INTO dataset_sources (
				dataset_id, source_key, name, url, description, publisher,
				source_type, licence_id, is_official, last_verified_at
			) VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''),
				'source', NULLIF($7, ''), false, $8)`,
			datasetID, sourceKey, sourceName, itemSource.URL, itemSource.Purpose,
			itemSource.Organization, item.LicenseID, accessedAt); err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO dataset_versions (
			dataset_id, version, format, status, record_count, storage_path,
			released_at
		) VALUES ($1, $2, $3, 'published', $4, $5, $6)`,
		datasetID, version, format, item.RecordCount, item.RelativePath, verifiedAt)
	return err
}

func nullableDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &parsed
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer(" ", "-", "/", "-", "&", "and", ",", "", ".", "").Replace(value)
	return value
}
