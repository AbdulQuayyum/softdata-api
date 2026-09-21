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
	nemaZonalTerritorialOperationOfficeRecordCount = 17
	nemaZonalTerritorialOperationOfficeDefaultPage = 1
	nemaZonalTerritorialOperationOfficeDefaultSize = 50
	nemaZonalTerritorialOperationOfficeMaxPageSize = 100
	nemaZonalTerritorialOperationOfficeMaxSearch   = 100
	nemaZonalTerritorialOperationOfficeType        = "zonal_territorial_operation_office"
)

var nemaZonalTerritorialOperationOfficeIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type NEMAZonalTerritorialOperationOfficeFileRepository struct {
	jsonRepository interfaces.JSONFileRepository
	recordsPath    string
	cache          lazyDatasetCache[nemaZonalTerritorialOperationOfficeDataset]
}

type nemaZonalTerritorialOperationOfficeDataset struct {
	records    []models.NEMAZonalTerritorialOperationOffice
	all        []int
	searchText []string
	byID       map[string]int
	byStateID  map[string][]int
	byType     map[string][]int
}

var _ interfaces.NEMAZonalTerritorialOperationOfficeRepository = (*NEMAZonalTerritorialOperationOfficeFileRepository)(nil)

func NewNEMAZonalTerritorialOperationOfficeRepository(jsonRepository interfaces.JSONFileRepository, recordsPath string) (*NEMAZonalTerritorialOperationOfficeFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanRecordsPath, err := validateGeographyDatasetPath("nema zonal territorial operation offices", recordsPath)
	if err != nil {
		return nil, err
	}
	return &NEMAZonalTerritorialOperationOfficeFileRepository{jsonRepository: jsonRepository, recordsPath: cleanRecordsPath}, nil
}

func (r *NEMAZonalTerritorialOperationOfficeFileRepository) ListNEMAZonalTerritorialOperationOffices(ctx context.Context, query interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeListResult, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return interfaces.NEMAZonalTerritorialOperationOfficeListResult{}, err
	}
	query, err = normalizeNEMAZonalTerritorialOperationOfficeQuery(query)
	if err != nil {
		return interfaces.NEMAZonalTerritorialOperationOfficeListResult{}, err
	}
	if err := contextError(ctx); err != nil {
		return interfaces.NEMAZonalTerritorialOperationOfficeListResult{}, err
	}

	candidates := dataset.all
	if query.StateID != "" {
		candidates = dataset.byStateID[query.StateID]
	} else if query.OfficeType != "" {
		candidates = dataset.byType[query.OfficeType]
	}

	matching := make([]int, 0, len(candidates))
	search := strings.ToLower(query.Search)
	for _, index := range candidates {
		if err := contextError(ctx); err != nil {
			return interfaces.NEMAZonalTerritorialOperationOfficeListResult{}, err
		}
		record := dataset.records[index]
		if query.StateID != "" && record.StateID != query.StateID {
			continue
		}
		if query.OfficeType != "" && record.OfficeType != query.OfficeType {
			continue
		}
		if search != "" && !strings.Contains(dataset.searchText[index], search) {
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
		return interfaces.NEMAZonalTerritorialOperationOfficeListResult{Records: make([]models.NEMAZonalTerritorialOperationOffice, 0), Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	page := make([]models.NEMAZonalTerritorialOperationOffice, end-start)
	for i, index := range matching[start:end] {
		page[i] = dataset.records[index]
	}
	return interfaces.NEMAZonalTerritorialOperationOfficeListResult{Records: page, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (r *NEMAZonalTerritorialOperationOfficeFileRepository) GetNEMAZonalTerritorialOperationOffice(ctx context.Context, id string) (models.NEMAZonalTerritorialOperationOffice, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return models.NEMAZonalTerritorialOperationOffice{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > models.NEMAZonalTerritorialOperationOfficeIDMaxLength || !nemaZonalTerritorialOperationOfficeIDPattern.MatchString(id) {
		return models.NEMAZonalTerritorialOperationOffice{}, fmt.Errorf("%w", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeQuery)
	}
	if err := contextError(ctx); err != nil {
		return models.NEMAZonalTerritorialOperationOffice{}, err
	}
	index, ok := dataset.byID[id]
	if !ok {
		return models.NEMAZonalTerritorialOperationOffice{}, fmt.Errorf("%w", interfaces.ErrNEMAZonalTerritorialOperationOfficeNotFound)
	}
	return dataset.records[index], nil
}

func (r *NEMAZonalTerritorialOperationOfficeFileRepository) load(ctx context.Context) (nemaZonalTerritorialOperationOfficeDataset, error) {
	values, err := r.cache.get(ctx, func(ctx context.Context) ([]nemaZonalTerritorialOperationOfficeDataset, error) {
		dataset, err := r.decodeAndIndex(ctx)
		if err != nil {
			return nil, err
		}
		return []nemaZonalTerritorialOperationOfficeDataset{dataset}, nil
	}, func(value []nemaZonalTerritorialOperationOfficeDataset) []nemaZonalTerritorialOperationOfficeDataset {
		return value
	})
	if err != nil {
		return nemaZonalTerritorialOperationOfficeDataset{}, err
	}
	return values[0], nil
}

func (r *NEMAZonalTerritorialOperationOfficeFileRepository) decodeAndIndex(ctx context.Context) (nemaZonalTerritorialOperationOfficeDataset, error) {
	var records []models.NEMAZonalTerritorialOperationOffice
	if err := r.jsonRepository.Decode(ctx, r.recordsPath, &records); err != nil {
		return nemaZonalTerritorialOperationOfficeDataset{}, sanitizeNEMAZonalTerritorialOperationOfficeLoadError(err)
	}
	if err := validateNEMAZonalTerritorialOperationOfficeRecords(ctx, records); err != nil {
		return nemaZonalTerritorialOperationOfficeDataset{}, err
	}
	dataset := nemaZonalTerritorialOperationOfficeDataset{
		records:    append([]models.NEMAZonalTerritorialOperationOffice(nil), records...),
		all:        make([]int, len(records)),
		searchText: make([]string, len(records)),
		byID:       make(map[string]int, len(records)),
		byStateID:  map[string][]int{},
		byType:     map[string][]int{},
	}
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return nemaZonalTerritorialOperationOfficeDataset{}, err
		}
		dataset.byID[record.ID] = i
		dataset.all[i] = i
		dataset.searchText[i] = strings.ToLower(record.Name)
		dataset.byStateID[record.StateID] = append(dataset.byStateID[record.StateID], i)
		dataset.byType[record.OfficeType] = append(dataset.byType[record.OfficeType], i)
	}
	return dataset, nil
}

func validateNEMAZonalTerritorialOperationOfficeRecords(ctx context.Context, records []models.NEMAZonalTerritorialOperationOffice) error {
	if len(records) != nemaZonalTerritorialOperationOfficeRecordCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	seenIDs := make(map[string]struct{}, len(records))
	seenStates := make(map[string]struct{}, len(records))
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return err
		}
		if record.ID == "" || strings.TrimSpace(record.Name) == "" || record.OfficeType != nemaZonalTerritorialOperationOfficeType || record.StateID == "" || record.CountryCode != "NG" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		for _, value := range []string{record.ID, record.Name, record.OfficeType, record.StateID, record.CountryCode} {
			if strings.TrimSpace(value) != value {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if len(record.ID) > models.NEMAZonalTerritorialOperationOfficeIDMaxLength || !nemaZonalTerritorialOperationOfficeIDPattern.MatchString(record.ID) || !validNEMAZonalTerritorialOperationOfficeState(record.StateID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenIDs[record.ID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[record.ID] = struct{}{}
		if _, duplicate := seenStates[record.StateID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenStates[record.StateID] = struct{}{}
		if i > 0 && records[i-1].ID > record.ID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	return nil
}

func normalizeNEMAZonalTerritorialOperationOfficeQuery(query interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeQuery, error) {
	if query.Page == 0 {
		query.Page = nemaZonalTerritorialOperationOfficeDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = nemaZonalTerritorialOperationOfficeDefaultSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > nemaZonalTerritorialOperationOfficeMaxPageSize {
		return interfaces.NEMAZonalTerritorialOperationOfficeQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeQuery)
	}
	rawStateID, rawOfficeType, rawSearch := query.StateID, query.OfficeType, query.Search
	query.StateID = strings.TrimSpace(query.StateID)
	query.OfficeType = strings.TrimSpace(query.OfficeType)
	query.Search = strings.TrimSpace(query.Search)
	if rawStateID != "" && query.StateID == "" {
		return interfaces.NEMAZonalTerritorialOperationOfficeQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeStateFilter)
	}
	if rawOfficeType != "" && query.OfficeType == "" {
		return interfaces.NEMAZonalTerritorialOperationOfficeQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeTypeFilter)
	}
	if rawSearch != "" && query.Search == "" {
		return interfaces.NEMAZonalTerritorialOperationOfficeQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeSearch)
	}
	if query.StateID != "" && !validNEMAZonalTerritorialOperationOfficeState(query.StateID) {
		return interfaces.NEMAZonalTerritorialOperationOfficeQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeStateFilter)
	}
	if query.OfficeType != "" && query.OfficeType != nemaZonalTerritorialOperationOfficeType {
		return interfaces.NEMAZonalTerritorialOperationOfficeQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeTypeFilter)
	}
	if len([]rune(query.Search)) > nemaZonalTerritorialOperationOfficeMaxSearch || strings.ContainsAny(query.Search, "\r\n\x00") {
		return interfaces.NEMAZonalTerritorialOperationOfficeQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeSearch)
	}
	return query, nil
}

func validNEMAZonalTerritorialOperationOfficeState(value string) bool {
	switch value {
	case "abia", "adamawa", "akwa-ibom", "anambra", "bauchi", "bayelsa", "benue", "borno", "cross-river", "delta", "ebonyi", "edo", "ekiti", "enugu", "fct", "gombe", "imo", "jigawa", "kaduna", "kano", "katsina", "kebbi", "kogi", "kwara", "lagos", "nasarawa", "niger", "ogun", "ondo", "osun", "oyo", "plateau", "rivers", "sokoto", "taraba", "yobe", "zamfara":
		return true
	default:
		return false
	}
}

func sanitizeNEMAZonalTerritorialOperationOfficeLoadError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, interfaces.ErrDatasetFileNotFound) || errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
}
