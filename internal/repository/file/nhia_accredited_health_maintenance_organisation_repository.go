package file

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	nhiaAccreditedHMORecordCount = 94
	nhiaAccreditedHMODefaultPage = 1
	nhiaAccreditedHMODefaultSize = 50
	nhiaAccreditedHMOMaxPageSize = 100
	nhiaAccreditedHMOMaxSearch   = 100
)

var nhiaAccreditedHMOIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NHIAAccreditedHealthMaintenanceOrganisationFileRepository provides lazy, indexed access to the NHIA accredited HMO snapshot.
type NHIAAccreditedHealthMaintenanceOrganisationFileRepository struct {
	jsonRepository interfaces.JSONFileRepository
	recordsPath    string
	cache          lazyDatasetCache[nhiaAccreditedHMODataset]
}

type nhiaAccreditedHMODataset struct {
	records     []models.NHIAAccreditedHealthMaintenanceOrganisation
	all         []int
	searchNames []string
	byID        map[string]int
	byHMOID     map[string][]int
	byStatus    map[string][]int
}

var _ interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository = (*NHIAAccreditedHealthMaintenanceOrganisationFileRepository)(nil)

// NewNHIAAccreditedHealthMaintenanceOrganisationRepository constructs a file-backed NHIA accredited HMO repository.
func NewNHIAAccreditedHealthMaintenanceOrganisationRepository(jsonRepository interfaces.JSONFileRepository, recordsPath string) (*NHIAAccreditedHealthMaintenanceOrganisationFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanRecordsPath, err := validateGeographyDatasetPath("nhia accredited health maintenance organisations", recordsPath)
	if err != nil {
		return nil, err
	}
	return &NHIAAccreditedHealthMaintenanceOrganisationFileRepository{jsonRepository: jsonRepository, recordsPath: cleanRecordsPath}, nil
}

func (r *NHIAAccreditedHealthMaintenanceOrganisationFileRepository) ListNHIAAccreditedHealthMaintenanceOrganisations(ctx context.Context, query interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{}, err
	}
	query, err = normalizeNHIAAccreditedHMOQuery(query)
	if err != nil {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{}, err
	}
	if err := contextError(ctx); err != nil {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{}, err
	}

	candidates := nhiaAccreditedHMOCandidates(dataset, query)
	matching := make([]int, 0, len(candidates))
	search := strings.ToLower(query.Search)
	for _, index := range candidates {
		if err := contextError(ctx); err != nil {
			return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{}, err
		}
		record := dataset.records[index]
		if query.AccreditationStatus != "" && record.AccreditationStatus != query.AccreditationStatus {
			continue
		}
		if query.HMOID != "" && record.HMOID != query.HMOID {
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
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{Records: make([]models.NHIAAccreditedHealthMaintenanceOrganisation, 0), Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	page := make([]models.NHIAAccreditedHealthMaintenanceOrganisation, end-start)
	for i, index := range matching[start:end] {
		page[i] = dataset.records[index]
	}
	return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{Records: page, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (r *NHIAAccreditedHealthMaintenanceOrganisationFileRepository) GetNHIAAccreditedHealthMaintenanceOrganisation(ctx context.Context, id string) (models.NHIAAccreditedHealthMaintenanceOrganisation, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, err
	}
	if err := contextError(ctx); err != nil {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, err
	}
	index, ok := dataset.byID[id]
	if !ok {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, fmt.Errorf("%w", interfaces.ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound)
	}
	return dataset.records[index], nil
}

func (r *NHIAAccreditedHealthMaintenanceOrganisationFileRepository) load(ctx context.Context) (nhiaAccreditedHMODataset, error) {
	values, err := r.cache.get(ctx, func(ctx context.Context) ([]nhiaAccreditedHMODataset, error) {
		dataset, err := r.decodeAndIndex(ctx)
		if err != nil {
			return nil, err
		}
		return []nhiaAccreditedHMODataset{dataset}, nil
	}, func(value []nhiaAccreditedHMODataset) []nhiaAccreditedHMODataset {
		return value
	})
	if err != nil {
		return nhiaAccreditedHMODataset{}, err
	}
	return values[0], nil
}

func (r *NHIAAccreditedHealthMaintenanceOrganisationFileRepository) decodeAndIndex(ctx context.Context) (nhiaAccreditedHMODataset, error) {
	var records []models.NHIAAccreditedHealthMaintenanceOrganisation
	if err := r.jsonRepository.Decode(ctx, r.recordsPath, &records); err != nil {
		return nhiaAccreditedHMODataset{}, sanitizeNHIAAccreditedHMOLoadError(err)
	}
	if err := validateNHIAAccreditedHMORecords(ctx, records); err != nil {
		return nhiaAccreditedHMODataset{}, err
	}

	dataset := nhiaAccreditedHMODataset{
		records:     records,
		all:         make([]int, len(records)),
		searchNames: make([]string, len(records)),
		byID:        make(map[string]int, len(records)),
		byHMOID:     map[string][]int{},
		byStatus:    map[string][]int{},
	}
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return nhiaAccreditedHMODataset{}, err
		}
		dataset.byID[record.ID] = i
		dataset.all[i] = i
		dataset.searchNames[i] = strings.ToLower(record.Name)
		dataset.byHMOID[record.HMOID] = append(dataset.byHMOID[record.HMOID], i)
		dataset.byStatus[record.AccreditationStatus] = append(dataset.byStatus[record.AccreditationStatus], i)
	}
	return dataset, nil
}

func validateNHIAAccreditedHMORecords(ctx context.Context, records []models.NHIAAccreditedHealthMaintenanceOrganisation) error {
	if len(records) != nhiaAccreditedHMORecordCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	seenIDs := make(map[string]struct{}, len(records))
	seenHMOIDs := make(map[string]struct{}, len(records))
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return err
		}
		if record.ID == "" || record.Name == "" || record.CountryCode != "NG" || record.OrganisationType != "health_maintenance_organisation" || record.AccreditationStatus != "accredited" || record.HMOID == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if len(record.ID) > models.NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength || !nhiaAccreditedHMOIDPattern.MatchString(record.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		for _, value := range []string{record.ID, record.Name, record.CountryCode, record.OrganisationType, record.AccreditationStatus, record.HMOID} {
			if strings.TrimSpace(value) != value {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if containsNHIAAccreditedHMOForbiddenText(record) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenIDs[record.ID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenHMOIDs[record.HMOID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[record.ID] = struct{}{}
		seenHMOIDs[record.HMOID] = struct{}{}
		if i > 0 && nhiaAccreditedHMOSortKey(records[i-1]) > nhiaAccreditedHMOSortKey(record) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	return nil
}

func normalizeNHIAAccreditedHMOQuery(query interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery, error) {
	if query.Page == 0 {
		query.Page = nhiaAccreditedHMODefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = nhiaAccreditedHMODefaultSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > nhiaAccreditedHMOMaxPageSize {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationQuery)
	}
	rawStatus := query.AccreditationStatus
	rawHMOID := query.HMOID
	rawSearch := query.Search
	query.AccreditationStatus = strings.TrimSpace(query.AccreditationStatus)
	query.HMOID = strings.TrimSpace(query.HMOID)
	query.Search = strings.TrimSpace(query.Search)
	if rawStatus != "" && query.AccreditationStatus == "" {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationQuery)
	}
	if rawHMOID != "" && query.HMOID == "" {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOIDFilter)
	}
	if rawSearch != "" && query.Search == "" {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch)
	}
	if query.AccreditationStatus != "" && query.AccreditationStatus != "accredited" {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatusFilter)
	}
	if strings.ContainsAny(query.HMOID, "\r\n\x00") {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOIDFilter)
	}
	if len([]rune(query.Search)) > nhiaAccreditedHMOMaxSearch || strings.ContainsAny(query.Search, "\r\n\x00") {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch)
	}
	return query, nil
}

func nhiaAccreditedHMOCandidates(dataset nhiaAccreditedHMODataset, query interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery) []int {
	sets := make([][]int, 0, 2)
	if query.AccreditationStatus != "" {
		sets = append(sets, dataset.byStatus[query.AccreditationStatus])
	}
	if query.HMOID != "" {
		sets = append(sets, dataset.byHMOID[query.HMOID])
	}
	if len(sets) == 0 {
		return dataset.all
	}
	sort.SliceStable(sets, func(i, j int) bool { return len(sets[i]) < len(sets[j]) })
	return sets[0]
}

func nhiaAccreditedHMOSortKey(record models.NHIAAccreditedHealthMaintenanceOrganisation) string {
	return strings.ToLower(record.Name) + "\x00" + nhiaAccreditedHMOOrderingHMOID(record.HMOID) + "\x00" + record.ID
}

func nhiaAccreditedHMOOrderingHMOID(hmoID string) string {
	if hmoID == "" {
		return hmoID
	}
	for _, char := range hmoID {
		if char < '0' || char > '9' {
			return hmoID
		}
	}
	if len(hmoID) >= 20 {
		return hmoID
	}
	return strings.Repeat("0", 20-len(hmoID)) + hmoID
}

func containsNHIAAccreditedHMOForbiddenText(record models.NHIAAccreditedHealthMaintenanceOrganisation) bool {
	combined := strings.ToLower(record.ID + "\x00" + record.Name + "\x00" + record.HMOID)
	for _, marker := range []string{"website", "logo", "phone", "email", "director", "contact person", "practitioner", "patient", "enrollee", "policy number", "registration status", "licence status", "operational status"} {
		if strings.Contains(combined, marker) {
			return true
		}
	}
	return false
}

func sanitizeNHIAAccreditedHMOLoadError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, interfaces.ErrDatasetFileNotFound) || errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
}
