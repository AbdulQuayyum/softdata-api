package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	NEMAZonalTerritorialOperationOfficeDefaultPage     = 1
	NEMAZonalTerritorialOperationOfficeDefaultPageSize = 50
	NEMAZonalTerritorialOperationOfficeMaxPageSize     = 100
	NEMAZonalTerritorialOperationOfficeMaxSearchLength = 100
)

var nemaZonalTerritorialOperationOfficeServiceIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type NEMAZonalTerritorialOperationOfficeQuery = interfaces.NEMAZonalTerritorialOperationOfficeQuery
type NEMAZonalTerritorialOperationOfficeListResult = interfaces.NEMAZonalTerritorialOperationOfficeListResult

type NEMAZonalTerritorialOperationOfficeService struct {
	repository interfaces.NEMAZonalTerritorialOperationOfficeRepository
}

func NewNEMAZonalTerritorialOperationOfficeService(repository interfaces.NEMAZonalTerritorialOperationOfficeRepository) (*NEMAZonalTerritorialOperationOfficeService, error) {
	if repository == nil {
		return nil, fmt.Errorf("nema zonal territorial operation office repository is required")
	}
	return &NEMAZonalTerritorialOperationOfficeService{repository: repository}, nil
}

func (s *NEMAZonalTerritorialOperationOfficeService) ListNEMAZonalTerritorialOperationOffices(ctx context.Context, input NEMAZonalTerritorialOperationOfficeQuery) (NEMAZonalTerritorialOperationOfficeListResult, error) {
	if err := serviceContextError(ctx); err != nil {
		return NEMAZonalTerritorialOperationOfficeListResult{}, err
	}
	query, err := normalizeNEMAZonalTerritorialOperationOfficeServiceQuery(input)
	if err != nil {
		return NEMAZonalTerritorialOperationOfficeListResult{}, err
	}
	result, err := s.repository.ListNEMAZonalTerritorialOperationOffices(ctx, query)
	if err != nil {
		return NEMAZonalTerritorialOperationOfficeListResult{}, translateNEMAZonalTerritorialOperationOfficeError("list nema zonal territorial operation offices", err)
	}
	result.Records = cloneNEMAZonalTerritorialOperationOfficeList(result.Records)
	return result, nil
}

func (s *NEMAZonalTerritorialOperationOfficeService) GetNEMAZonalTerritorialOperationOffice(ctx context.Context, id string) (models.NEMAZonalTerritorialOperationOffice, error) {
	if err := serviceContextError(ctx); err != nil {
		return models.NEMAZonalTerritorialOperationOffice{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > models.NEMAZonalTerritorialOperationOfficeIDMaxLength || !nemaZonalTerritorialOperationOfficeServiceIDPattern.MatchString(id) {
		return models.NEMAZonalTerritorialOperationOffice{}, ErrInvalidNEMAZonalTerritorialOperationOfficeID
	}
	record, err := s.repository.GetNEMAZonalTerritorialOperationOffice(ctx, id)
	if err != nil {
		return models.NEMAZonalTerritorialOperationOffice{}, translateNEMAZonalTerritorialOperationOfficeError("get nema zonal territorial operation office", err)
	}
	return record, nil
}

func normalizeNEMAZonalTerritorialOperationOfficeServiceQuery(input NEMAZonalTerritorialOperationOfficeQuery) (NEMAZonalTerritorialOperationOfficeQuery, error) {
	query := input
	if query.Page == 0 {
		query.Page = NEMAZonalTerritorialOperationOfficeDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = NEMAZonalTerritorialOperationOfficeDefaultPageSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > NEMAZonalTerritorialOperationOfficeMaxPageSize {
		return NEMAZonalTerritorialOperationOfficeQuery{}, ErrInvalidNEMAZonalTerritorialOperationOfficePagination
	}
	rawStateID, rawOfficeType, rawSearch := query.StateID, query.OfficeType, query.Search
	query.StateID = strings.TrimSpace(query.StateID)
	query.OfficeType = strings.TrimSpace(query.OfficeType)
	query.Search = strings.TrimSpace(query.Search)
	if rawStateID != "" && query.StateID == "" {
		return NEMAZonalTerritorialOperationOfficeQuery{}, ErrInvalidNEMAZonalTerritorialOperationOfficeStateID
	}
	if rawOfficeType != "" && query.OfficeType == "" {
		return NEMAZonalTerritorialOperationOfficeQuery{}, ErrInvalidNEMAZonalTerritorialOperationOfficeType
	}
	if rawSearch != "" && query.Search == "" {
		return NEMAZonalTerritorialOperationOfficeQuery{}, ErrInvalidNEMAZonalTerritorialOperationOfficeSearch
	}
	if query.StateID != "" && !validNEMAZonalTerritorialOperationOfficeStateID(query.StateID) {
		return NEMAZonalTerritorialOperationOfficeQuery{}, ErrInvalidNEMAZonalTerritorialOperationOfficeStateID
	}
	if query.OfficeType != "" && query.OfficeType != "zonal_territorial_operation_office" {
		return NEMAZonalTerritorialOperationOfficeQuery{}, ErrInvalidNEMAZonalTerritorialOperationOfficeType
	}
	if len([]rune(query.Search)) > NEMAZonalTerritorialOperationOfficeMaxSearchLength || strings.ContainsAny(query.Search, "\r\n\x00") {
		return NEMAZonalTerritorialOperationOfficeQuery{}, ErrInvalidNEMAZonalTerritorialOperationOfficeSearch
	}
	return query, nil
}

func translateNEMAZonalTerritorialOperationOfficeError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrNEMAZonalTerritorialOperationOfficeNotFound):
		return ErrNEMAZonalTerritorialOperationOfficeNotFound
	case errors.Is(err, interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeQuery):
		return ErrInvalidNEMAZonalTerritorialOperationOfficePagination
	case errors.Is(err, interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeStateFilter):
		return ErrInvalidNEMAZonalTerritorialOperationOfficeStateID
	case errors.Is(err, interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeTypeFilter):
		return ErrInvalidNEMAZonalTerritorialOperationOfficeType
	case errors.Is(err, interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeSearch):
		return ErrInvalidNEMAZonalTerritorialOperationOfficeSearch
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return ErrInvalidNEMAZonalTerritorialOperationOfficeDataset
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func cloneNEMAZonalTerritorialOperationOfficeList(items []models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice {
	if len(items) == 0 {
		return make([]models.NEMAZonalTerritorialOperationOffice, 0)
	}
	cloned := make([]models.NEMAZonalTerritorialOperationOffice, len(items))
	copy(cloned, items)
	return cloned
}

func validNEMAZonalTerritorialOperationOfficeStateID(value string) bool {
	switch value {
	case "abia", "adamawa", "akwa-ibom", "anambra", "bauchi", "bayelsa", "benue", "borno", "cross-river", "delta", "ebonyi", "edo", "ekiti", "enugu", "fct", "gombe", "imo", "jigawa", "kaduna", "kano", "katsina", "kebbi", "kogi", "kwara", "lagos", "nasarawa", "niger", "ogun", "ondo", "osun", "oyo", "plateau", "rivers", "sokoto", "taraba", "yobe", "zamfara":
		return true
	default:
		return false
	}
}
