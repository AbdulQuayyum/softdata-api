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
	NHIAStateSocialHealthInsuranceAgencyDefaultPage     = 1
	NHIAStateSocialHealthInsuranceAgencyDefaultPageSize = 50
	NHIAStateSocialHealthInsuranceAgencyMaxPageSize     = 100
	NHIAStateSocialHealthInsuranceAgencyMaxSearchLength = 100
)

var nhiaSSHIAServiceIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type NHIAStateSocialHealthInsuranceAgencyQuery = interfaces.NHIAStateSocialHealthInsuranceAgencyQuery
type NHIAStateSocialHealthInsuranceAgencyListResult = interfaces.NHIAStateSocialHealthInsuranceAgencyListResult

type NHIAStateSocialHealthInsuranceAgencyService struct {
	repository interfaces.NHIAStateSocialHealthInsuranceAgencyRepository
}

func NewNHIAStateSocialHealthInsuranceAgencyService(repository interfaces.NHIAStateSocialHealthInsuranceAgencyRepository) (*NHIAStateSocialHealthInsuranceAgencyService, error) {
	if repository == nil {
		return nil, fmt.Errorf("nhia state social health insurance agency repository is required")
	}
	return &NHIAStateSocialHealthInsuranceAgencyService{repository: repository}, nil
}

func (s *NHIAStateSocialHealthInsuranceAgencyService) ListNHIAStateSocialHealthInsuranceAgencies(ctx context.Context, input NHIAStateSocialHealthInsuranceAgencyQuery) (NHIAStateSocialHealthInsuranceAgencyListResult, error) {
	if err := serviceContextError(ctx); err != nil {
		return NHIAStateSocialHealthInsuranceAgencyListResult{}, err
	}
	query, err := normalizeNHIASSHIAServiceQuery(input)
	if err != nil {
		return NHIAStateSocialHealthInsuranceAgencyListResult{}, err
	}
	result, err := s.repository.ListNHIAStateSocialHealthInsuranceAgencies(ctx, query)
	if err != nil {
		return NHIAStateSocialHealthInsuranceAgencyListResult{}, translateNHIASSHIAError("list NHIA state social health insurance agencies", err)
	}
	result.Records = cloneNHIASSHIAList(result.Records)
	return result, nil
}

func (s *NHIAStateSocialHealthInsuranceAgencyService) GetNHIAStateSocialHealthInsuranceAgency(ctx context.Context, id string) (models.NHIAStateSocialHealthInsuranceAgency, error) {
	if err := serviceContextError(ctx); err != nil {
		return models.NHIAStateSocialHealthInsuranceAgency{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > models.NHIAStateSocialHealthInsuranceAgencyIDMaxLength || !nhiaSSHIAServiceIDPattern.MatchString(id) {
		return models.NHIAStateSocialHealthInsuranceAgency{}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyID
	}
	record, err := s.repository.GetNHIAStateSocialHealthInsuranceAgency(ctx, id)
	if err != nil {
		return models.NHIAStateSocialHealthInsuranceAgency{}, translateNHIASSHIAError("get NHIA state social health insurance agency", err)
	}
	return record, nil
}

func normalizeNHIASSHIAServiceQuery(input NHIAStateSocialHealthInsuranceAgencyQuery) (NHIAStateSocialHealthInsuranceAgencyQuery, error) {
	query := input
	if query.Page == 0 {
		query.Page = NHIAStateSocialHealthInsuranceAgencyDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = NHIAStateSocialHealthInsuranceAgencyDefaultPageSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > NHIAStateSocialHealthInsuranceAgencyMaxPageSize {
		return NHIAStateSocialHealthInsuranceAgencyQuery{}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyPagination
	}
	rawStateID := query.StateID
	rawSearch := query.Search
	query.StateID = strings.TrimSpace(query.StateID)
	query.Search = strings.TrimSpace(query.Search)
	if rawStateID != "" && query.StateID == "" {
		return NHIAStateSocialHealthInsuranceAgencyQuery{}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateID
	}
	if rawSearch != "" && query.Search == "" {
		return NHIAStateSocialHealthInsuranceAgencyQuery{}, ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch
	}
	if query.StateID != "" && (!nhiaSSHIAServiceIDPattern.MatchString(query.StateID) || !validNHIASSHIAServiceStateID(query.StateID)) {
		return NHIAStateSocialHealthInsuranceAgencyQuery{}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateID
	}
	if len([]rune(query.Search)) > NHIAStateSocialHealthInsuranceAgencyMaxSearchLength || strings.ContainsAny(query.Search, "\r\n\x00") {
		return NHIAStateSocialHealthInsuranceAgencyQuery{}, ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch
	}
	return query, nil
}

func translateNHIASSHIAError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrNHIAStateSocialHealthInsuranceAgencyNotFound):
		return ErrNHIAStateSocialHealthInsuranceAgencyNotFound
	case errors.Is(err, interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencyQuery):
		return ErrInvalidNHIAStateSocialHealthInsuranceAgencyPagination
	case errors.Is(err, interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateFilter):
		return ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateID
	case errors.Is(err, interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch):
		return ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return ErrInvalidNHIAStateSocialHealthInsuranceAgencyDataset
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func cloneNHIASSHIAList(items []models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency {
	if len(items) == 0 {
		return make([]models.NHIAStateSocialHealthInsuranceAgency, 0)
	}
	cloned := make([]models.NHIAStateSocialHealthInsuranceAgency, len(items))
	copy(cloned, items)
	return cloned
}

func validNHIASSHIAServiceStateID(stateID string) bool {
	_, ok := nhiaSSHIAServiceStateIDs[stateID]
	return ok
}

var nhiaSSHIAServiceStateIDs = map[string]struct{}{
	"abia": {}, "adamawa": {}, "akwa-ibom": {}, "anambra": {}, "bauchi": {},
	"bayelsa": {}, "benue": {}, "borno": {}, "cross-river": {}, "delta": {},
	"ebonyi": {}, "edo": {}, "ekiti": {}, "enugu": {}, "fct": {}, "gombe": {},
	"imo": {}, "jigawa": {}, "kaduna": {}, "kano": {}, "katsina": {},
	"kebbi": {}, "kogi": {}, "kwara": {}, "lagos": {}, "nasarawa": {},
	"niger": {}, "ogun": {}, "ondo": {}, "osun": {}, "oyo": {}, "plateau": {},
	"rivers": {}, "sokoto": {}, "taraba": {}, "yobe": {}, "zamfara": {},
}
