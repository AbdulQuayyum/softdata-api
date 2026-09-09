package file

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	medicalLaboratoryAccreditationRecordCount = 30
	medicalLaboratoryAccreditationAccredited  = 26
	medicalLaboratoryAccreditationExpired     = 4
	medicalLaboratoryAccreditationDefaultPage = 1
	medicalLaboratoryAccreditationDefaultSize = 50
	medicalLaboratoryAccreditationMaxPageSize = 100
	medicalLaboratoryAccreditationMaxSearch   = 100
)

var medicalLaboratoryAccreditationIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var medicalLaboratoryAccreditationNumberPattern = regexp.MustCompile(`^ML[0-9]{4}$`)

var medicalLaboratoryAccreditationStatuses = map[string]struct{}{
	"accredited": {},
	"expired":    {},
}

// MedicalLaboratoryAccreditationFileRepository provides lazy, indexed access to the MLSCN accreditation snapshot.
type MedicalLaboratoryAccreditationFileRepository struct {
	jsonRepository interfaces.JSONFileRepository
	recordsPath    string
	statesPath     string
	cache          lazyDatasetCache[medicalLaboratoryAccreditationDataset]
}

type medicalLaboratoryAccreditationDataset struct {
	records     []models.MedicalLaboratoryAccreditation
	all         []int
	searchNames []string
	byID        map[string]int
	byState     map[string][]int
	byStatus    map[string][]int
	stateIDs    map[string]struct{}
}

var _ interfaces.MedicalLaboratoryAccreditationRepository = (*MedicalLaboratoryAccreditationFileRepository)(nil)

// NewMedicalLaboratoryAccreditationRepository constructs a file-backed accreditation repository.
func NewMedicalLaboratoryAccreditationRepository(jsonRepository interfaces.JSONFileRepository, recordsPath, statesPath string) (*MedicalLaboratoryAccreditationFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanRecordsPath, err := validateGeographyDatasetPath("medical laboratory accreditations", recordsPath)
	if err != nil {
		return nil, err
	}
	cleanStatesPath, err := validateGeographyDatasetPath("states", statesPath)
	if err != nil {
		return nil, err
	}
	return &MedicalLaboratoryAccreditationFileRepository{jsonRepository: jsonRepository, recordsPath: cleanRecordsPath, statesPath: cleanStatesPath}, nil
}

func (r *MedicalLaboratoryAccreditationFileRepository) ListMedicalLaboratoryAccreditations(ctx context.Context, query interfaces.MedicalLaboratoryAccreditationQuery) (interfaces.MedicalLaboratoryAccreditationListResult, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return interfaces.MedicalLaboratoryAccreditationListResult{}, err
	}
	query, err = normalizeMedicalLaboratoryAccreditationQuery(query, dataset)
	if err != nil {
		return interfaces.MedicalLaboratoryAccreditationListResult{}, err
	}
	if err := contextError(ctx); err != nil {
		return interfaces.MedicalLaboratoryAccreditationListResult{}, err
	}

	candidates := medicalLaboratoryAccreditationCandidates(dataset, query)
	matching := make([]int, 0, len(candidates))
	search := strings.ToLower(query.Search)
	for _, index := range candidates {
		if err := contextError(ctx); err != nil {
			return interfaces.MedicalLaboratoryAccreditationListResult{}, err
		}
		record := dataset.records[index]
		if query.StateID != "" && record.StateID != query.StateID {
			continue
		}
		if query.AccreditationStatus != "" && record.AccreditationStatus != query.AccreditationStatus {
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
		return interfaces.MedicalLaboratoryAccreditationListResult{Records: make([]models.MedicalLaboratoryAccreditation, 0), Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	page := make([]models.MedicalLaboratoryAccreditation, end-start)
	for i, index := range matching[start:end] {
		page[i] = dataset.records[index]
	}
	return interfaces.MedicalLaboratoryAccreditationListResult{Records: page, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (r *MedicalLaboratoryAccreditationFileRepository) GetMedicalLaboratoryAccreditation(ctx context.Context, id string) (models.MedicalLaboratoryAccreditation, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return models.MedicalLaboratoryAccreditation{}, err
	}
	if err := contextError(ctx); err != nil {
		return models.MedicalLaboratoryAccreditation{}, err
	}
	index, ok := dataset.byID[id]
	if !ok {
		return models.MedicalLaboratoryAccreditation{}, fmt.Errorf("%w", interfaces.ErrMedicalLaboratoryAccreditationNotFound)
	}
	return dataset.records[index], nil
}

func (r *MedicalLaboratoryAccreditationFileRepository) load(ctx context.Context) (medicalLaboratoryAccreditationDataset, error) {
	values, err := r.cache.get(ctx, func(ctx context.Context) ([]medicalLaboratoryAccreditationDataset, error) {
		dataset, err := r.decodeAndIndex(ctx)
		if err != nil {
			return nil, err
		}
		return []medicalLaboratoryAccreditationDataset{dataset}, nil
	}, func(value []medicalLaboratoryAccreditationDataset) []medicalLaboratoryAccreditationDataset {
		return value
	})
	if err != nil {
		return medicalLaboratoryAccreditationDataset{}, err
	}
	return values[0], nil
}

func (r *MedicalLaboratoryAccreditationFileRepository) decodeAndIndex(ctx context.Context) (medicalLaboratoryAccreditationDataset, error) {
	var records []models.MedicalLaboratoryAccreditation
	if err := r.jsonRepository.Decode(ctx, r.recordsPath, &records); err != nil {
		return medicalLaboratoryAccreditationDataset{}, sanitizeMedicalLaboratoryAccreditationLoadError(err)
	}
	var states []models.State
	if err := r.jsonRepository.Decode(ctx, r.statesPath, &states); err != nil {
		return medicalLaboratoryAccreditationDataset{}, sanitizeMedicalLaboratoryAccreditationLoadError(err)
	}

	stateIDs := make(map[string]struct{}, len(states))
	for _, state := range states {
		if err := contextError(ctx); err != nil {
			return medicalLaboratoryAccreditationDataset{}, err
		}
		if state.ID == "" || state.CountryCode != "NG" || !medicalLaboratoryAccreditationIDPattern.MatchString(state.ID) {
			return medicalLaboratoryAccreditationDataset{}, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		stateIDs[state.ID] = struct{}{}
	}
	if err := validateMedicalLaboratoryAccreditationRecords(ctx, records, stateIDs); err != nil {
		return medicalLaboratoryAccreditationDataset{}, err
	}

	dataset := medicalLaboratoryAccreditationDataset{records: records, all: make([]int, len(records)), searchNames: make([]string, len(records)), byID: make(map[string]int, len(records)), byState: map[string][]int{}, byStatus: map[string][]int{}, stateIDs: stateIDs}
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return medicalLaboratoryAccreditationDataset{}, err
		}
		dataset.byID[record.ID] = i
		dataset.all[i] = i
		dataset.searchNames[i] = strings.ToLower(record.Name)
		dataset.byState[record.StateID] = append(dataset.byState[record.StateID], i)
		dataset.byStatus[record.AccreditationStatus] = append(dataset.byStatus[record.AccreditationStatus], i)
	}
	return dataset, nil
}

func validateMedicalLaboratoryAccreditationRecords(ctx context.Context, records []models.MedicalLaboratoryAccreditation, stateIDs map[string]struct{}) error {
	if len(records) != medicalLaboratoryAccreditationRecordCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	seenIDs := make(map[string]struct{}, len(records))
	seenAccreditationNumbers := make(map[string]struct{}, len(records))
	statusCounts := map[string]int{"accredited": 0, "expired": 0}
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return err
		}
		if record.ID == "" || record.Name == "" || record.StateID == "" || record.CountryCode != "NG" || record.AccreditationStatus == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if len(record.ID) > models.MedicalLaboratoryAccreditationIDMaxLength || !medicalLaboratoryAccreditationIDPattern.MatchString(record.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if strings.TrimSpace(record.Name) != record.Name || strings.TrimSpace(record.Address) != record.Address || strings.TrimSpace(record.AccreditationNumber) != record.AccreditationNumber {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := stateIDs[record.StateID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := medicalLaboratoryAccreditationStatuses[record.AccreditationStatus]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		statusCounts[record.AccreditationStatus]++
		if record.AccreditationNumber != "" {
			if !medicalLaboratoryAccreditationNumberPattern.MatchString(record.AccreditationNumber) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, duplicate := seenAccreditationNumbers[record.AccreditationNumber]; duplicate {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			seenAccreditationNumbers[record.AccreditationNumber] = struct{}{}
		}
		for _, dateValue := range []string{record.ApprovalDate, record.ExpiryDate} {
			if dateValue == "" {
				continue
			}
			if strings.TrimSpace(dateValue) != dateValue {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, err := time.Parse("2006-01-02", dateValue); err != nil {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if containsMedicalLaboratoryAccreditationForbiddenText(record) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenIDs[record.ID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[record.ID] = struct{}{}
		if i > 0 && medicalLaboratoryAccreditationSortKey(records[i-1]) > medicalLaboratoryAccreditationSortKey(record) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	if statusCounts["accredited"] != medicalLaboratoryAccreditationAccredited || statusCounts["expired"] != medicalLaboratoryAccreditationExpired {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func normalizeMedicalLaboratoryAccreditationQuery(query interfaces.MedicalLaboratoryAccreditationQuery, dataset medicalLaboratoryAccreditationDataset) (interfaces.MedicalLaboratoryAccreditationQuery, error) {
	if query.Page == 0 {
		query.Page = medicalLaboratoryAccreditationDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = medicalLaboratoryAccreditationDefaultSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > medicalLaboratoryAccreditationMaxPageSize {
		return interfaces.MedicalLaboratoryAccreditationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidMedicalLaboratoryAccreditationQuery)
	}
	query.StateID = strings.TrimSpace(query.StateID)
	query.AccreditationStatus = strings.TrimSpace(query.AccreditationStatus)
	query.Search = strings.TrimSpace(query.Search)
	if query.StateID != "" {
		if _, ok := dataset.stateIDs[query.StateID]; !ok || !medicalLaboratoryAccreditationIDPattern.MatchString(query.StateID) {
			return interfaces.MedicalLaboratoryAccreditationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidMedicalLaboratoryAccreditationStateFilter)
		}
	}
	if query.AccreditationStatus != "" {
		if _, ok := medicalLaboratoryAccreditationStatuses[query.AccreditationStatus]; !ok {
			return interfaces.MedicalLaboratoryAccreditationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidMedicalLaboratoryAccreditationStatusFilter)
		}
	}
	if len([]rune(query.Search)) > medicalLaboratoryAccreditationMaxSearch || strings.ContainsAny(query.Search, "\r\n\x00") {
		return interfaces.MedicalLaboratoryAccreditationQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidMedicalLaboratoryAccreditationSearch)
	}
	return query, nil
}

func medicalLaboratoryAccreditationCandidates(dataset medicalLaboratoryAccreditationDataset, query interfaces.MedicalLaboratoryAccreditationQuery) []int {
	sets := make([][]int, 0, 2)
	if query.StateID != "" {
		sets = append(sets, dataset.byState[query.StateID])
	}
	if query.AccreditationStatus != "" {
		sets = append(sets, dataset.byStatus[query.AccreditationStatus])
	}
	if len(sets) == 0 {
		return dataset.all
	}
	sort.SliceStable(sets, func(i, j int) bool { return len(sets[i]) < len(sets[j]) })
	return sets[0]
}

func medicalLaboratoryAccreditationSortKey(record models.MedicalLaboratoryAccreditation) string {
	return record.StateID + "\x00" + strings.ToLower(record.Name) + "\x00" + record.ID
}

func containsMedicalLaboratoryAccreditationForbiddenText(record models.MedicalLaboratoryAccreditation) bool {
	combined := strings.ToLower(record.Name + "\x00" + record.Address + "\x00" + record.AccreditationNumber)
	for _, marker := range []string{"superintendent", "phone", "email", "personal registration", "practitioner record"} {
		if strings.Contains(combined, marker) {
			return true
		}
	}
	return false
}

func sanitizeMedicalLaboratoryAccreditationLoadError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, interfaces.ErrDatasetFileNotFound) || errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
}
