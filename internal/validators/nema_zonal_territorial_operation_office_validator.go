package validators

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var nemaZonalTerritorialOperationOfficeSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidateNEMAZonalTerritorialOperationOfficeID(field, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError(field, "Office ID is required.")
	}
	if strings.Contains(value, "/") || strings.Contains(value, "\\") || strings.ContainsAny(value, "?#") {
		return "", invalidField(field, "Office ID must be a valid lowercase public slug.")
	}
	if len(value) > models.NEMAZonalTerritorialOperationOfficeIDMaxLength || !nemaZonalTerritorialOperationOfficeSlugPattern.MatchString(value) {
		return "", invalidField(field, "Office ID must be a valid lowercase public slug.")
	}
	return value, nil
}

func ValidateNEMAZonalTerritorialOperationOfficeListQuery(values url.Values) (interfaces.NEMAZonalTerritorialOperationOfficeQuery, error) {
	var errs ValidationErrors
	query := interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 50}
	allowed := map[string]struct{}{"page": {}, "page_size": {}, "state_id": {}, "office_type": {}, "search": {}}
	for field := range values {
		if _, ok := allowed[field]; !ok {
			errs.Add(field, codeInvalid, "Query parameter is not supported.")
		}
	}
	readSingle := func(field string) (string, bool) {
		items := values[field]
		if len(items) > 1 {
			errs.Add(field, codeMalformed, "Field may be provided at most once.")
			return "", false
		}
		if len(items) == 0 {
			return "", false
		}
		return items[0], true
	}
	parsePositive := func(field, raw string, defaultValue, maximum int) int {
		value, err := strconv.ParseUint(raw, 10, 0)
		if err != nil || value > uint64(^uint(0)>>1) || value < 1 || (maximum > 0 && int(value) > maximum) {
			errs.Add(field, codeInvalid, "Value must be a valid positive integer within the supported range.")
			return defaultValue
		}
		return int(value)
	}
	if raw, ok := readSingle("page"); ok {
		query.Page = parsePositive("page", raw, 1, 0)
	}
	if raw, ok := readSingle("page_size"); ok {
		query.PageSize = parsePositive("page_size", raw, 50, 100)
	}
	if raw, ok := readSingle("state_id"); ok {
		query.StateID = strings.TrimSpace(raw)
		if query.StateID == "" {
			errs.Add("state_id", codeRequired, "State ID is required when provided.")
		} else if !nemaZonalTerritorialOperationOfficeSlugPattern.MatchString(query.StateID) || !validNEMAOfficeStateID(query.StateID) {
			errs.Add("state_id", codeInvalid, "State ID must be a canonical Nigerian state or FCT slug.")
		}
	}
	if raw, ok := readSingle("office_type"); ok {
		query.OfficeType = strings.TrimSpace(raw)
		if query.OfficeType == "" {
			errs.Add("office_type", codeRequired, "Office type is required when provided.")
		} else if query.OfficeType != "zonal_territorial_operation_office" {
			errs.Add("office_type", codeInvalid, "Value is not supported.")
		}
	}
	if raw, ok := readSingle("search"); ok {
		query.Search = strings.TrimSpace(raw)
		if query.Search == "" {
			query.Search = ""
		} else if len([]rune(query.Search)) > 100 || strings.ContainsAny(query.Search, "\r\n\x00") {
			errs.Add("search", codeInvalid, "Search must be at most 100 characters.")
			query.Search = ""
		}
	}
	if len(errs.Fields) > 0 {
		return interfaces.NEMAZonalTerritorialOperationOfficeQuery{}, errs
	}
	return query, nil
}

func validNEMAOfficeStateID(stateID string) bool {
	_, ok := nemaOfficeStateIDs[stateID]
	return ok
}

var nemaOfficeStateIDs = map[string]struct{}{
	"abia": {}, "adamawa": {}, "akwa-ibom": {}, "anambra": {}, "bauchi": {},
	"bayelsa": {}, "benue": {}, "borno": {}, "cross-river": {}, "delta": {},
	"ebonyi": {}, "edo": {}, "ekiti": {}, "enugu": {}, "fct": {}, "gombe": {},
	"imo": {}, "jigawa": {}, "kaduna": {}, "kano": {}, "katsina": {},
	"kebbi": {}, "kogi": {}, "kwara": {}, "lagos": {}, "nasarawa": {},
	"niger": {}, "ogun": {}, "ondo": {}, "osun": {}, "oyo": {}, "plateau": {},
	"rivers": {}, "sokoto": {}, "taraba": {}, "yobe": {}, "zamfara": {},
}
