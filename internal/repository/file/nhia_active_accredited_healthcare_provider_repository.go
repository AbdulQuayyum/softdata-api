package file

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	nhiaHCPRecordCount = 6536
	nhiaHCPDefaultPage = 1
	nhiaHCPDefaultSize = 50
	nhiaHCPMaxPageSize = 100
	nhiaHCPMaxSearch   = 100
)

var (
	nhiaHCPIDPattern           = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	nhiaHCPProviderCodePattern = regexp.MustCompile(`^(?:[A-Z]{2,3})/[0-9]{4}/P$`)
)

type NHIAActiveAccreditedHealthcareProviderFileRepository struct {
	jsonRepository interfaces.JSONFileRepository
	recordsPath    string
	cache          lazyDatasetCache[nhiaHCPDataset]
}

type nhiaHCPDataset struct {
	records         []models.NHIAActiveAccreditedHealthcareProvider
	all             []int
	searchNames     []string
	byID            map[string]int
	byProviderCode  map[string][]int
	byFacilityType  map[string][]int
	byListingStatus map[string][]int
}

var _ interfaces.NHIAActiveAccreditedHealthcareProviderRepository = (*NHIAActiveAccreditedHealthcareProviderFileRepository)(nil)

func NewNHIAActiveAccreditedHealthcareProviderRepository(jsonRepository interfaces.JSONFileRepository, recordsPath string) (*NHIAActiveAccreditedHealthcareProviderFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanRecordsPath, err := validateGeographyDatasetPath("nhia active accredited healthcare providers", recordsPath)
	if err != nil {
		return nil, err
	}
	return &NHIAActiveAccreditedHealthcareProviderFileRepository{jsonRepository: jsonRepository, recordsPath: cleanRecordsPath}, nil
}

func (r *NHIAActiveAccreditedHealthcareProviderFileRepository) ListNHIAActiveAccreditedHealthcareProviders(ctx context.Context, query interfaces.NHIAActiveAccreditedHealthcareProviderQuery) (interfaces.NHIAActiveAccreditedHealthcareProviderListResult, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{}, err
	}
	query, err = normalizeNHIAHCPQuery(query)
	if err != nil {
		return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{}, err
	}
	if err := contextError(ctx); err != nil {
		return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{}, err
	}

	candidates := dataset.all
	if query.ProviderCode != "" {
		candidates = dataset.byProviderCode[query.ProviderCode]
	} else if query.FacilityType != "" {
		candidates = dataset.byFacilityType[query.FacilityType]
	} else if query.ListingStatus != "" {
		candidates = dataset.byListingStatus[query.ListingStatus]
	}

	matching := make([]int, 0, len(candidates))
	search := strings.ToLower(query.Search)
	for _, index := range candidates {
		if err := contextError(ctx); err != nil {
			return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{}, err
		}
		record := dataset.records[index]
		if query.ProviderCode != "" && record.ProviderCode != query.ProviderCode {
			continue
		}
		if query.FacilityType != "" && record.FacilityType != query.FacilityType {
			continue
		}
		if query.ListingStatus != "" && record.ListingStatus != query.ListingStatus {
			continue
		}
		if search != "" && !strings.Contains(dataset.searchNames[index], search) {
			continue
		}
		matching = append(matching, index)
	}

	total := len(matching)
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}
	start := 0
	if query.Page > 1 && query.Page-1 > math.MaxInt/query.PageSize {
		start = total
	} else {
		start = (query.Page - 1) * query.PageSize
	}
	if start >= total {
		return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{Records: make([]models.NHIAActiveAccreditedHealthcareProvider, 0), Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	page := make([]models.NHIAActiveAccreditedHealthcareProvider, end-start)
	for i, index := range matching[start:end] {
		page[i] = dataset.records[index]
	}
	return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{Records: page, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (r *NHIAActiveAccreditedHealthcareProviderFileRepository) GetNHIAActiveAccreditedHealthcareProvider(ctx context.Context, id string) (models.NHIAActiveAccreditedHealthcareProvider, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return models.NHIAActiveAccreditedHealthcareProvider{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > models.NHIAActiveAccreditedHealthcareProviderIDMaxLength || !nhiaHCPIDPattern.MatchString(id) {
		return models.NHIAActiveAccreditedHealthcareProvider{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderQuery)
	}
	if err := contextError(ctx); err != nil {
		return models.NHIAActiveAccreditedHealthcareProvider{}, err
	}
	index, ok := dataset.byID[id]
	if !ok {
		return models.NHIAActiveAccreditedHealthcareProvider{}, fmt.Errorf("%w", interfaces.ErrNHIAActiveAccreditedHealthcareProviderNotFound)
	}
	return dataset.records[index], nil
}

func (r *NHIAActiveAccreditedHealthcareProviderFileRepository) load(ctx context.Context) (nhiaHCPDataset, error) {
	values, err := r.cache.get(ctx, func(ctx context.Context) ([]nhiaHCPDataset, error) {
		dataset, err := r.decodeAndIndex(ctx)
		if err != nil {
			return nil, err
		}
		return []nhiaHCPDataset{dataset}, nil
	}, func(value []nhiaHCPDataset) []nhiaHCPDataset { return value })
	if err != nil {
		return nhiaHCPDataset{}, err
	}
	return values[0], nil
}

func (r *NHIAActiveAccreditedHealthcareProviderFileRepository) decodeAndIndex(ctx context.Context) (nhiaHCPDataset, error) {
	var records []models.NHIAActiveAccreditedHealthcareProvider
	if err := r.jsonRepository.Decode(ctx, r.recordsPath, &records); err != nil {
		return nhiaHCPDataset{}, sanitizeNHIAHCPLoadError(err)
	}
	if err := validateNHIAHCPRecords(ctx, records); err != nil {
		return nhiaHCPDataset{}, err
	}
	dataset := nhiaHCPDataset{
		records:         append([]models.NHIAActiveAccreditedHealthcareProvider(nil), records...),
		all:             make([]int, len(records)),
		searchNames:     make([]string, len(records)),
		byID:            make(map[string]int, len(records)),
		byProviderCode:  map[string][]int{},
		byFacilityType:  map[string][]int{},
		byListingStatus: map[string][]int{},
	}
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return nhiaHCPDataset{}, err
		}
		dataset.byID[record.ID] = i
		dataset.all[i] = i
		dataset.searchNames[i] = strings.ToLower(record.Name)
		dataset.byProviderCode[record.ProviderCode] = append(dataset.byProviderCode[record.ProviderCode], i)
		dataset.byFacilityType[record.FacilityType] = append(dataset.byFacilityType[record.FacilityType], i)
		dataset.byListingStatus[record.ListingStatus] = append(dataset.byListingStatus[record.ListingStatus], i)
	}
	return dataset, nil
}

func validateNHIAHCPRecords(ctx context.Context, records []models.NHIAActiveAccreditedHealthcareProvider) error {
	if len(records) != nhiaHCPRecordCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	seenIDs := make(map[string]struct{}, len(records))
	seenCodes := make(map[string]struct{}, len(records))
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return err
		}
		if record.ID == "" || record.Name == "" || record.CountryCode != "NG" || record.ProviderCode == "" || record.FacilityType == "" || record.ListingStatus == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		for _, value := range []string{record.ID, record.Name, record.CountryCode, record.ProviderCode, record.FacilityType, record.ListingStatus} {
			if strings.TrimSpace(value) != value {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if len(record.ID) > models.NHIAActiveAccreditedHealthcareProviderIDMaxLength || !nhiaHCPIDPattern.MatchString(record.ID) || !nhiaHCPProviderCodePattern.MatchString(record.ProviderCode) || !validNHIAHCPFacilityType(record.FacilityType) || !validNHIAHCPListingStatus(record.ListingStatus) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenIDs[record.ID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenCodes[record.ProviderCode]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[record.ID] = struct{}{}
		seenCodes[record.ProviderCode] = struct{}{}
		if i > 0 && records[i-1].ID > record.ID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	return nil
}

func normalizeNHIAHCPQuery(query interfaces.NHIAActiveAccreditedHealthcareProviderQuery) (interfaces.NHIAActiveAccreditedHealthcareProviderQuery, error) {
	if query.Page == 0 {
		query.Page = nhiaHCPDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = nhiaHCPDefaultSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > nhiaHCPMaxPageSize {
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderQuery)
	}
	rawCode, rawType, rawStatus, rawSearch := query.ProviderCode, query.FacilityType, query.ListingStatus, query.Search
	query.ProviderCode = strings.TrimSpace(query.ProviderCode)
	query.FacilityType = strings.TrimSpace(query.FacilityType)
	query.ListingStatus = strings.TrimSpace(query.ListingStatus)
	query.Search = strings.TrimSpace(query.Search)
	if rawCode != "" && query.ProviderCode == "" {
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderCodeFilter)
	}
	if rawType != "" && query.FacilityType == "" {
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderFacilityTypeFilter)
	}
	if rawStatus != "" && query.ListingStatus == "" {
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderListingStatusFilter)
	}
	if rawSearch != "" && query.Search == "" {
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderSearch)
	}
	if query.ProviderCode != "" && !nhiaHCPProviderCodePattern.MatchString(query.ProviderCode) {
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderCodeFilter)
	}
	if query.FacilityType != "" && !validNHIAHCPFacilityType(query.FacilityType) {
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderFacilityTypeFilter)
	}
	if query.ListingStatus != "" && !validNHIAHCPListingStatus(query.ListingStatus) {
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderListingStatusFilter)
	}
	if len([]rune(query.Search)) > nhiaHCPMaxSearch || strings.ContainsAny(query.Search, "\r\n\x00") {
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderSearch)
	}
	return query, nil
}

func validNHIAHCPFacilityType(value string) bool {
	return value == "primary" || value == "primary_and_secondary"
}

func validNHIAHCPListingStatus(value string) bool {
	return value == "active_accredited"
}

func sanitizeNHIAHCPLoadError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, interfaces.ErrDatasetFileNotFound) || errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
}
