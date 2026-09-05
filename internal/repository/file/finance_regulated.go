package file

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/datasets/assets"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var regulatedIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
var regulatedCBNPattern = regexp.MustCompile(`^[0-9]{3}$`)
var regulatedNIPPattern = regexp.MustCompile(`^[0-9]{6}$`)

func (r *FinanceFileRepository) ListNonInterestFinancialInstitutions(ctx context.Context) ([]models.NonInterestInstitution, error) {
	rows, err := r.loadRegulated(ctx, r.nonInterestFinancialInstitutionsPath, "non-interest", 6, true, interfaces.ErrNonInterestFinancialInstitutionNotFound, []models.NonInterestInstitution{})
	if err != nil {
		return nil, err
	}
	return rows.([]models.NonInterestInstitution), nil
}

func (r *FinanceFileRepository) GetNonInterestFinancialInstitution(ctx context.Context, id string) (models.NonInterestInstitution, error) {
	rows, err := r.ListNonInterestFinancialInstitutions(ctx)
	return getRegulated(rows, id, interfaces.ErrNonInterestFinancialInstitutionNotFound, err)
}

func (r *FinanceFileRepository) ListMerchantBanks(ctx context.Context) ([]models.MerchantBank, error) {
	rows, err := r.loadRegulated(ctx, r.merchantBanksPath, "merchant-banks", 6, true, interfaces.ErrMerchantBankNotFound, []models.MerchantBank{})
	if err != nil {
		return nil, err
	}
	return rows.([]models.MerchantBank), nil
}
func (r *FinanceFileRepository) GetMerchantBank(ctx context.Context, id string) (models.MerchantBank, error) {
	rows, err := r.ListMerchantBanks(ctx)
	return getRegulated(rows, id, interfaces.ErrMerchantBankNotFound, err)
}

func (r *FinanceFileRepository) ListPaymentServiceBanks(ctx context.Context) ([]models.PaymentServiceBank, error) {
	rows, err := r.loadRegulated(ctx, r.paymentServiceBanksPath, "payment-service-banks", 5, true, interfaces.ErrPaymentServiceBankNotFound, []models.PaymentServiceBank{})
	if err != nil {
		return nil, err
	}
	return rows.([]models.PaymentServiceBank), nil
}
func (r *FinanceFileRepository) GetPaymentServiceBank(ctx context.Context, id string) (models.PaymentServiceBank, error) {
	rows, err := r.ListPaymentServiceBanks(ctx)
	return getRegulated(rows, id, interfaces.ErrPaymentServiceBankNotFound, err)
}

func (r *FinanceFileRepository) ListFinancialHoldingCompanies(ctx context.Context) ([]models.FinancialHoldingCompany, error) {
	rows, err := r.loadRegulated(ctx, r.financialHoldingCompaniesPath, "holding-companies", 7, false, interfaces.ErrFinancialHoldingCompanyNotFound, []models.FinancialHoldingCompany{})
	if err != nil {
		return nil, err
	}
	return rows.([]models.FinancialHoldingCompany), nil
}
func (r *FinanceFileRepository) GetFinancialHoldingCompany(ctx context.Context, id string) (models.FinancialHoldingCompany, error) {
	rows, err := r.ListFinancialHoldingCompanies(ctx)
	return getRegulated(rows, id, interfaces.ErrFinancialHoldingCompanyNotFound, err)
}

func (r *FinanceFileRepository) ListDevelopmentFinanceInstitutions(ctx context.Context) ([]models.DevelopmentFinanceInstitution, error) {
	rows, err := r.loadRegulated(ctx, r.developmentFinanceInstitutionsPath, "development-finance", 8, true, interfaces.ErrDevelopmentFinanceInstitutionNotFound, []models.DevelopmentFinanceInstitution{})
	if err != nil {
		return nil, err
	}
	return rows.([]models.DevelopmentFinanceInstitution), nil
}
func (r *FinanceFileRepository) GetDevelopmentFinanceInstitution(ctx context.Context, id string) (models.DevelopmentFinanceInstitution, error) {
	rows, err := r.ListDevelopmentFinanceInstitutions(ctx)
	return getRegulated(rows, id, interfaces.ErrDevelopmentFinanceInstitutionNotFound, err)
}

func (r *FinanceFileRepository) ListPrimaryMortgageInstitutions(ctx context.Context) ([]models.PrimaryMortgageInstitution, error) {
	rows, err := r.loadRegulated(ctx, r.primaryMortgageInstitutionsPath, "primary-mortgage", 31, true, interfaces.ErrPrimaryMortgageInstitutionNotFound, []models.PrimaryMortgageInstitution{})
	if err != nil {
		return nil, err
	}
	return rows.([]models.PrimaryMortgageInstitution), nil
}
func (r *FinanceFileRepository) GetPrimaryMortgageInstitution(ctx context.Context, id string) (models.PrimaryMortgageInstitution, error) {
	rows, err := r.ListPrimaryMortgageInstitutions(ctx)
	return getRegulated(rows, id, interfaces.ErrPrimaryMortgageInstitutionNotFound, err)
}

func (r *FinanceFileRepository) loadRegulated(ctx context.Context, path, category string, count int, bankCodes bool, notFound error, target any) (any, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r == nil || r.jsonRepository == nil || strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	cacheKey := path + "|" + reflect.TypeOf(target).String()
	r.regulatedMu.RLock()
	cached := r.regulatedCache[cacheKey]
	r.regulatedMu.RUnlock()
	if cached != nil {
		return cloneRegulated(cached), nil
	}
	raw := make([]json.RawMessage, 0)
	if err := r.jsonRepository.Decode(ctx, path, &raw); err != nil {
		return nil, translateFinanceLoadError(err)
	}
	if len(raw) != count {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	result := reflect.New(reflect.TypeOf(target)).Elem()
	result.Set(reflect.MakeSlice(result.Type(), len(raw), len(raw)))
	for i, item := range raw {
		fields := map[string]json.RawMessage{}
		if err := json.Unmarshal(item, &fields); err != nil {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		allowed := map[string]bool{"id": true, "name": true, "country_code": true, "website_url": true, "logo_url": true, "cbn_code": bankCodes, "nip_code": bankCodes}
		for key := range fields {
			if !allowed[key] {
				return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if err := json.Unmarshal(item, result.Index(i).Addr().Interface()); err != nil {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	value := result.Interface()
	if err := validateRegulated(value, category, count, bankCodes); err != nil {
		return nil, err
	}
	r.regulatedMu.Lock()
	if r.regulatedCache == nil {
		r.regulatedCache = make(map[string]any)
	}
	r.regulatedCache[cacheKey] = value
	r.regulatedMu.Unlock()
	return cloneRegulated(value), nil
}

func validateRegulated(value any, category string, count int, bankCodes bool) error {
	v := reflect.ValueOf(value)
	if v.Len() != count {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	ids, names, cbns, nips, logos := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	previousName, previousID := "", ""
	seenPMI := map[string]bool{}
	for i := 0; i < v.Len(); i++ {
		row := v.Index(i)
		id := row.FieldByName("ID").String()
		name := row.FieldByName("Name").String()
		country := row.FieldByName("CountryCode").String()
		lowerName := strings.ToLower(name)
		if !regulatedIDPattern.MatchString(id) || name == "" || country != "NG" || ids[id] || names[name] || (previousName != "" && (strings.ToLower(previousName) > lowerName || (strings.ToLower(previousName) == lowerName && previousID > id))) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		ids[id], names[name], previousName, previousID = true, true, name, id
		if category == "primary-mortgage" {
			seenPMI[id] = true
		}
		for _, field := range []string{"WebsiteURL", "LogoURL"} {
			f := row.FieldByName(field).String()
			if f == "" {
				continue
			}
			if field == "WebsiteURL" {
				u, err := url.Parse(f)
				if err != nil || u.Scheme != "https" || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
					return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
				}
			} else {
				expected := "/v1/assets/financial-institutions/ng/" + category + "/" + id + ".png"
				if f != expected || logos[f] {
					return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
				}
				if _, err := assets.FinancialInstitutionLogo(category, id, "png"); err != nil {
					return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
				}
				logos[f] = true
			}
		}
		if bankCodes {
			for field, spec := range map[string]struct {
				p *regexp.Regexp
				s map[string]bool
			}{"CBNCode": {regulatedCBNPattern, cbns}, "NIPCode": {regulatedNIPPattern, nips}} {
				f := row.FieldByName(field).String()
				if f != "" {
					if !spec.p.MatchString(f) || spec.s[f] {
						return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
					}
					spec.s[f] = true
				}
			}
		}
	}
	if category == "primary-mortgage" && (!seenPMI["firsttrust-mortgage-bank"] || seenPMI["aso-savings-loans"] || seenPMI["trustbond-mortgage-bank"]) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func cloneRegulated(value any) any {
	v := reflect.ValueOf(value)
	out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
	reflect.Copy(out, v)
	return out.Interface()
}

func getRegulated[T any](rows []T, id string, notFound error, err error) (T, error) {
	var zero T
	if err != nil {
		return zero, err
	}
	id = strings.TrimSpace(id)
	if !regulatedIDPattern.MatchString(id) {
		return zero, fmt.Errorf("%w", notFound)
	}
	for _, row := range rows {
		if reflect.ValueOf(row).FieldByName("ID").String() == id {
			return row, nil
		}
	}
	return zero, fmt.Errorf("%w", notFound)
}
