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
	nhiaSSHIARecordCount = 37
	nhiaSSHIADefaultPage = 1
	nhiaSSHIADefaultSize = 50
	nhiaSSHIAMaxPageSize = 100
	nhiaSSHIAMaxSearch   = 100
)

var nhiaSSHIAIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type NHIAStateSocialHealthInsuranceAgencyFileRepository struct {
	jsonRepository interfaces.JSONFileRepository
	recordsPath    string
	cache          lazyDatasetCache[nhiaSSHIADataset]
}

type nhiaSSHIADataset struct {
	records     []models.NHIAStateSocialHealthInsuranceAgency
	all         []int
	searchNames []string
	byID        map[string]int
	byStateID   map[string][]int
}

var _ interfaces.NHIAStateSocialHealthInsuranceAgencyRepository = (*NHIAStateSocialHealthInsuranceAgencyFileRepository)(nil)

func NewNHIAStateSocialHealthInsuranceAgencyRepository(jsonRepository interfaces.JSONFileRepository, recordsPath string) (*NHIAStateSocialHealthInsuranceAgencyFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanRecordsPath, err := validateGeographyDatasetPath("nhia state social health insurance agencies", recordsPath)
	if err != nil {
		return nil, err
	}
	return &NHIAStateSocialHealthInsuranceAgencyFileRepository{jsonRepository: jsonRepository, recordsPath: cleanRecordsPath}, nil
}

func (r *NHIAStateSocialHealthInsuranceAgencyFileRepository) ListNHIAStateSocialHealthInsuranceAgencies(ctx context.Context, query interfaces.NHIAStateSocialHealthInsuranceAgencyQuery) (interfaces.NHIAStateSocialHealthInsuranceAgencyListResult, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{}, err
	}
	query, err = normalizeNHIASSHIAQuery(query)
	if err != nil {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{}, err
	}
	if err := contextError(ctx); err != nil {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{}, err
	}

	candidates := dataset.all
	if query.StateID != "" {
		candidates = dataset.byStateID[query.StateID]
	}
	matching := make([]int, 0, len(candidates))
	search := strings.ToLower(query.Search)
	for _, index := range candidates {
		if err := contextError(ctx); err != nil {
			return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{}, err
		}
		record := dataset.records[index]
		if query.StateID != "" && record.StateID != query.StateID {
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
		return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{Records: make([]models.NHIAStateSocialHealthInsuranceAgency, 0), Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	page := make([]models.NHIAStateSocialHealthInsuranceAgency, end-start)
	for i, index := range matching[start:end] {
		page[i] = dataset.records[index]
	}
	return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{Records: page, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (r *NHIAStateSocialHealthInsuranceAgencyFileRepository) GetNHIAStateSocialHealthInsuranceAgency(ctx context.Context, id string) (models.NHIAStateSocialHealthInsuranceAgency, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return models.NHIAStateSocialHealthInsuranceAgency{}, err
	}
	if err := contextError(ctx); err != nil {
		return models.NHIAStateSocialHealthInsuranceAgency{}, err
	}
	index, ok := dataset.byID[id]
	if !ok {
		return models.NHIAStateSocialHealthInsuranceAgency{}, fmt.Errorf("%w", interfaces.ErrNHIAStateSocialHealthInsuranceAgencyNotFound)
	}
	return dataset.records[index], nil
}

func (r *NHIAStateSocialHealthInsuranceAgencyFileRepository) load(ctx context.Context) (nhiaSSHIADataset, error) {
	values, err := r.cache.get(ctx, func(ctx context.Context) ([]nhiaSSHIADataset, error) {
		dataset, err := r.decodeAndIndex(ctx)
		if err != nil {
			return nil, err
		}
		return []nhiaSSHIADataset{dataset}, nil
	}, func(value []nhiaSSHIADataset) []nhiaSSHIADataset {
		return value
	})
	if err != nil {
		return nhiaSSHIADataset{}, err
	}
	return values[0], nil
}

func (r *NHIAStateSocialHealthInsuranceAgencyFileRepository) decodeAndIndex(ctx context.Context) (nhiaSSHIADataset, error) {
	var records []models.NHIAStateSocialHealthInsuranceAgency
	if err := r.jsonRepository.Decode(ctx, r.recordsPath, &records); err != nil {
		return nhiaSSHIADataset{}, sanitizeNHIASSHIALoadError(err)
	}
	if err := validateNHIASSHIARecords(ctx, records); err != nil {
		return nhiaSSHIADataset{}, err
	}
	dataset := nhiaSSHIADataset{
		records:     append([]models.NHIAStateSocialHealthInsuranceAgency(nil), records...),
		all:         make([]int, len(records)),
		searchNames: make([]string, len(records)),
		byID:        make(map[string]int, len(records)),
		byStateID:   map[string][]int{},
	}
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return nhiaSSHIADataset{}, err
		}
		dataset.byID[record.ID] = i
		dataset.all[i] = i
		dataset.searchNames[i] = strings.ToLower(record.Name)
		dataset.byStateID[record.StateID] = append(dataset.byStateID[record.StateID], i)
	}
	return dataset, nil
}

func validateNHIASSHIARecords(ctx context.Context, records []models.NHIAStateSocialHealthInsuranceAgency) error {
	if len(records) != nhiaSSHIARecordCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	seenIDs := make(map[string]struct{}, len(records))
	seenStates := make(map[string]struct{}, len(records))
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return err
		}
		if record.ID == "" || record.Name == "" || record.StateID == "" || record.CountryCode != "NG" || record.OrganisationType != "state_social_health_insurance_agency" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if len(record.ID) > models.NHIAStateSocialHealthInsuranceAgencyIDMaxLength || !nhiaSSHIAIDPattern.MatchString(record.ID) || !validNHIASSHIAStateID(record.StateID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		for _, value := range []string{record.ID, record.Name, record.StateID, record.CountryCode, record.OrganisationType} {
			if strings.TrimSpace(value) != value {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if containsNHIASSHIAForbiddenText(record) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenIDs[record.ID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenStates[record.StateID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[record.ID] = struct{}{}
		seenStates[record.StateID] = struct{}{}
		if i > 0 && records[i-1].StateID > record.StateID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	if len(seenStates) != len(nhiaSSHIACanonicalStateIDs) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	for stateID := range nhiaSSHIACanonicalStateIDs {
		if _, ok := seenStates[stateID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	return nil
}

func normalizeNHIASSHIAQuery(query interfaces.NHIAStateSocialHealthInsuranceAgencyQuery) (interfaces.NHIAStateSocialHealthInsuranceAgencyQuery, error) {
	if query.Page == 0 {
		query.Page = nhiaSSHIADefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = nhiaSSHIADefaultSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > nhiaSSHIAMaxPageSize {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencyQuery)
	}
	rawStateID := query.StateID
	rawSearch := query.Search
	query.StateID = strings.TrimSpace(query.StateID)
	query.Search = strings.TrimSpace(query.Search)
	if rawStateID != "" && query.StateID == "" {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateFilter)
	}
	if rawSearch != "" && query.Search == "" {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch)
	}
	if query.StateID != "" && (!nhiaSSHIAIDPattern.MatchString(query.StateID) || !validNHIASSHIAStateID(query.StateID)) {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateFilter)
	}
	if len([]rune(query.Search)) > nhiaSSHIAMaxSearch || strings.ContainsAny(query.Search, "\r\n\x00") {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch)
	}
	return query, nil
}

func validNHIASSHIAStateID(stateID string) bool {
	_, ok := nhiaSSHIACanonicalStateIDs[stateID]
	return ok
}

var nhiaSSHIACanonicalStateIDs = map[string]struct{}{
	"abia": {}, "adamawa": {}, "akwa-ibom": {}, "anambra": {}, "bauchi": {},
	"bayelsa": {}, "benue": {}, "borno": {}, "cross-river": {}, "delta": {},
	"ebonyi": {}, "edo": {}, "ekiti": {}, "enugu": {}, "fct": {}, "gombe": {},
	"imo": {}, "jigawa": {}, "kaduna": {}, "kano": {}, "katsina": {},
	"kebbi": {}, "kogi": {}, "kwara": {}, "lagos": {}, "nasarawa": {},
	"niger": {}, "ogun": {}, "ondo": {}, "osun": {}, "oyo": {}, "plateau": {},
	"rivers": {}, "sokoto": {}, "taraba": {}, "yobe": {}, "zamfara": {},
}

func containsNHIASSHIAForbiddenText(record models.NHIAStateSocialHealthInsuranceAgency) bool {
	combined := strings.ToLower(record.ID + "\x00" + record.Name + "\x00" + record.StateID)
	for _, marker := range []string{"website", "logo", "phone", "email", "director", "contact person", "address", "accreditation status", "registration status", "licence status", "license status", "operational status", "personal"} {
		if strings.Contains(combined, marker) {
			return true
		}
	}
	return false
}

func sanitizeNHIASSHIALoadError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, interfaces.ErrDatasetFileNotFound) || errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
}
