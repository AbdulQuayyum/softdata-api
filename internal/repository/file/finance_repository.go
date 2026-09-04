package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/datasets/assets"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var financePaymentServiceProviderIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
var financePaymentServiceProviderSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)
var financePaymentServiceProviderCollapsePattern = regexp.MustCompile(`-+`)

var financeInternationalMoneyTransferOperatorIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var financeInternationalMoneyTransferOperatorSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)
var financeInternationalMoneyTransferOperatorCollapsePattern = regexp.MustCompile(`-+`)

var financeCurrencyIDPattern = regexp.MustCompile(`^[a-z]{3}$`)
var financeCurrencyAlphabeticCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)
var financeCurrencyNumericCodePattern = regexp.MustCompile(`^[0-9]{3}$`)
var financeCurrencyCountryAreaIDPattern = regexp.MustCompile(`^[a-z]{2}$`)
var financeCommercialBankIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
var financeCommercialBankCBNCodePattern = regexp.MustCompile(`^[0-9]{3}$`)
var financeCommercialBankNIPCodePattern = regexp.MustCompile(`^[0-9]{6}$`)

var financePaymentServiceProviderTypeOrder = map[string]int{
	"mobile_money_operator":               0,
	"switching_and_processing_company":    1,
	"payment_solution_service_provider":   2,
	"payment_terminal_service_provider":   3,
	"super_agent":                         4,
	"payment_service_holding_company":     5,
	"payment_terminal_service_aggregator": 6,
}

var financeExpectedPaymentServiceProviderCounts = map[string]int{
	"mobile_money_operator":               17,
	"switching_and_processing_company":    19,
	"payment_solution_service_provider":   108,
	"payment_terminal_service_provider":   47,
	"super_agent":                         61,
	"payment_service_holding_company":     1,
	"payment_terminal_service_aggregator": 2,
}

var financeExpectedInternationalMoneyTransferOperatorCount = 108

var financeExpectedCurrencyCount = 155
var financeExpectedCurrencyRelationshipCount = 252
var financeExpectedCurrencyCountryAreaCount = 245
var financeCurrencyZeroMappingCountryAreaIDs = map[string]struct{}{
	"aq": {},
	"gs": {},
	"ps": {},
}

var financeCurrencyExcludedCodes = map[string]struct{}{
	"BOV": {}, "CHE": {}, "CHW": {}, "CLF": {}, "COU": {}, "MXV": {}, "USN": {}, "UYI": {}, "UYW": {},
	"XAD": {}, "XAG": {}, "XAU": {}, "XBA": {}, "XBB": {}, "XBC": {}, "XBD": {}, "XDR": {}, "XPD": {},
	"XPT": {}, "XSU": {}, "XTS": {}, "XUA": {}, "XXX": {},
}

const (
	financeCurrenciesRelativePath        = "finance/currencies.json"
	financeCountriesAndAreasRelativePath = "geography/countries_and_areas.json"
	financeCommercialBanksRelativePath   = "finance/commercial_banks.json"
)

var financeExpectedCommercialBankIDs = []string{
	"access-bank", "alpha-morgan-bank", "citibank-nigeria", "ecobank-nigeria", "fidelity-bank",
	"first-bank-of-nigeria", "first-city-monument-bank", "globus-bank", "guaranty-trust-bank", "keystone-bank",
	"nova-bank", "optimus-bank", "parallex-bank", "polaris-bank", "premium-trust-bank", "providus-bank",
	"signature-bank", "stanbic-ibtc-bank", "standard-chartered-bank", "sterling-bank", "suntrust-bank", "tatum-bank",
	"titan-trust-bank", "union-bank", "united-bank-for-africa", "unity-bank", "wema-bank", "zenith-bank",
}

var financeExpectedCommercialBankNames = map[string]string{
	"access-bank": "Access Bank Plc", "alpha-morgan-bank": "Alpha Morgan Bank Limited", "citibank-nigeria": "Citibank Nigeria Limited",
	"ecobank-nigeria": "Ecobank Nigeria Plc", "fidelity-bank": "Fidelity Bank Plc", "first-bank-of-nigeria": "First Bank of Nigeria Limited",
	"first-city-monument-bank": "First City Monument Bank Plc", "globus-bank": "Globus Bank Limited", "guaranty-trust-bank": "Guaranty Trust Bank Plc",
	"keystone-bank": "Keystone Bank Limited", "nova-bank": "Nova Commercial Bank Limited", "optimus-bank": "Optimus Bank",
	"parallex-bank": "Parallex Bank Ltd", "polaris-bank": "Polaris Bank Plc", "premium-trust-bank": "Premium Trust Bank",
	"providus-bank": "Providus Bank", "signature-bank": "Signature Bank Limited", "stanbic-ibtc-bank": "Stanbic IBTC Bank Plc",
	"standard-chartered-bank": "Standard Chartered Bank Nigeria Limited", "sterling-bank": "Sterling Bank Plc", "suntrust-bank": "SunTrust Bank Nigeria Limited",
	"tatum-bank": "Tatum Bank Limited", "titan-trust-bank": "Titan Trust Bank Ltd", "union-bank": "Union Bank of Nigeria Plc",
	"united-bank-for-africa": "United Bank for Africa Plc", "unity-bank": "Unity Bank Plc", "wema-bank": "Wema Bank Plc", "zenith-bank": "Zenith Bank Plc",
}

var financeInternationalMoneyTransferOperatorFormerNames = map[string]struct{}{
	"FLUTTERWAVE TECHNOLOGY SOLUTIONS LTD": {},
	"STERLING CURRENCY EXCHANGE LTD":       {},
	"MULTIGATE GROUP HOLDING LIMITED":      {},
	"PAGATECH LIMITED":                     {},
	"VTNETWORK LIMITED":                    {},
	"ALALAMIYA EXCHANGE LIMITED":           {},
	"FIEM GROUP LLC DBA":                   {},
}

// FinanceFileRepository reads payment-service-provider records from a JSON dataset file.
type FinanceFileRepository struct {
	jsonRepository                          interfaces.JSONFileRepository
	paymentServiceProvidersPath             string
	internationalMoneyTransferOperatorsPath string
	currenciesPath                          string
	countriesAndAreasPath                   string
	commercialBanksPath                     string
}

var _ interfaces.FinanceRepository = (*FinanceFileRepository)(nil)

// NewFinanceRepository constructs a file-backed finance repository.
func NewFinanceRepository(jsonRepository interfaces.JSONFileRepository, paymentServiceProvidersPath string, internationalMoneyTransferOperatorsPath ...string) (*FinanceFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanedPath, err := validateFinanceDatasetPath(paymentServiceProvidersPath)
	if err != nil {
		return nil, err
	}
	imtoPath := ""
	if len(internationalMoneyTransferOperatorsPath) > 1 {
		return nil, fmt.Errorf("international money transfer operators path accepts at most one value")
	}
	if len(internationalMoneyTransferOperatorsPath) == 1 {
		imtoPath, err = validateFinanceDatasetPath(internationalMoneyTransferOperatorsPath[0])
		if err != nil {
			return nil, err
		}
	}

	return &FinanceFileRepository{
		jsonRepository:                          jsonRepository,
		paymentServiceProvidersPath:             cleanedPath,
		internationalMoneyTransferOperatorsPath: imtoPath,
		currenciesPath:                          financeCurrenciesRelativePath,
		countriesAndAreasPath:                   financeCountriesAndAreasRelativePath,
		commercialBanksPath:                     financeCommercialBanksRelativePath,
	}, nil
}

// ListCommercialBanks returns the ordered list of CBN-listed commercial banks.
func (r *FinanceFileRepository) ListCommercialBanks(ctx context.Context) ([]models.CommercialBank, error) {
	banks, err := r.loadCommercialBanks(ctx)
	if err != nil {
		return nil, err
	}
	return cloneCommercialBankList(banks), nil
}

// GetCommercialBank returns one commercial bank using its exact public ID.
func (r *FinanceFileRepository) GetCommercialBank(ctx context.Context, bankID string) (models.CommercialBank, error) {
	bankID = strings.TrimSpace(bankID)
	if bankID == "" || !financeCommercialBankIDPattern.MatchString(bankID) {
		return models.CommercialBank{}, fmt.Errorf("%w", interfaces.ErrCommercialBankNotFound)
	}
	banks, err := r.loadCommercialBanks(ctx)
	if err != nil {
		return models.CommercialBank{}, err
	}
	for _, bank := range banks {
		if bank.ID == bankID {
			return bank, nil
		}
	}
	return models.CommercialBank{}, fmt.Errorf("%w", interfaces.ErrCommercialBankNotFound)
}

func (r *FinanceFileRepository) loadCommercialBanks(ctx context.Context) ([]models.CommercialBank, error) {
	if r == nil || r.jsonRepository == nil || strings.TrimSpace(r.commercialBanksPath) == "" {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var rawBanks []json.RawMessage
	if err := r.jsonRepository.Decode(ctx, r.commercialBanksPath, &rawBanks); err != nil {
		return nil, translateFinanceLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(rawBanks) != len(financeExpectedCommercialBankIDs) {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	banks := make([]models.CommercialBank, len(rawBanks))
	for i, raw := range rawBanks {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		for _, field := range []string{"cbn_code", "nip_code"} {
			if value, ok := fields[field]; ok && (bytesEqual(value, []byte("null")) || bytesEqual(value, []byte(`""`))) {
				return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if err := json.Unmarshal(raw, &banks[i]); err != nil {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	if err := validateCommercialBanks(banks); err != nil {
		return nil, err
	}
	return banks, nil
}

func validateCommercialBanks(banks []models.CommercialBank) error {
	if len(banks) != 28 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	seenIDs := make(map[string]struct{}, len(banks))
	seenNames := make(map[string]struct{}, len(banks))
	seenCBN := make(map[string]struct{}, len(banks))
	seenNIP := make(map[string]struct{}, len(banks))
	prevName, prevID := "", ""
	for _, bank := range banks {
		if bank.ID == "" || bank.Name == "" || bank.CountryCode != "NG" || bank.OfficialWebsiteURL == "" || bank.LogoURL == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !financeCommercialBankIDPattern.MatchString(bank.ID) || financeExpectedCommercialBankNames[bank.ID] != bank.Name {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[bank.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenNames[bank.Name]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		parsed, err := url.Parse(bank.OfficialWebsiteURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if bank.LogoURL != "/v1/assets/banks/ng/"+bank.ID+".png" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, err := assets.BankLogo(bank.ID, "png"); err != nil {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if bank.CBNCode != "" {
			if !financeCommercialBankCBNCodePattern.MatchString(bank.CBNCode) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, ok := seenCBN[bank.CBNCode]; ok {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			seenCBN[bank.CBNCode] = struct{}{}
		}
		if bank.NIPCode != "" {
			if !financeCommercialBankNIPCodePattern.MatchString(bank.NIPCode) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, ok := seenNIP[bank.NIPCode]; ok {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			seenNIP[bank.NIPCode] = struct{}{}
		}
		if prevName != "" && (strings.ToLower(prevName) > strings.ToLower(bank.Name) || (strings.EqualFold(prevName, bank.Name) && prevID > bank.ID)) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[bank.ID] = struct{}{}
		seenNames[bank.Name] = struct{}{}
		prevName, prevID = bank.Name, bank.ID
	}
	if len(seenIDs) != len(financeExpectedCommercialBankIDs) || len(seenCBN) != 25 || len(seenNIP) != 25 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if banks[10].ID != "nova-bank" || banks[10].CBNCode != "" || banks[10].NIPCode != "" || banks[1].CBNCode != "" || banks[16].CBNCode != "" || banks[18].NIPCode != "" || banks[20].NIPCode != "" {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func bytesEqual(left, right []byte) bool { return string(left) == string(right) }

// ListPaymentServiceProviders returns the ordered list of payment-service-provider memberships.
func (r *FinanceFileRepository) ListPaymentServiceProviders(ctx context.Context) ([]models.PaymentServiceProvider, error) {
	providers, err := r.loadPaymentServiceProviders(ctx)
	if err != nil {
		return nil, err
	}
	return clonePaymentServiceProviderList(providers), nil
}

// ListPaymentServiceProvidersByType returns the ordered list of memberships for one institution type.
func (r *FinanceFileRepository) ListPaymentServiceProvidersByType(ctx context.Context, institutionType string) ([]models.PaymentServiceProvider, error) {
	providers, err := r.loadPaymentServiceProviders(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]models.PaymentServiceProvider, 0, financeExpectedPaymentServiceProviderCounts[institutionType])
	for _, provider := range providers {
		if provider.InstitutionType == institutionType {
			filtered = append(filtered, provider)
		}
	}
	return clonePaymentServiceProviderList(filtered), nil
}

// GetPaymentServiceProvider returns a single payment-service-provider membership using its public slug identifier.
func (r *FinanceFileRepository) GetPaymentServiceProvider(ctx context.Context, providerID string) (models.PaymentServiceProvider, error) {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" || !financePaymentServiceProviderIDPattern.MatchString(providerID) {
		return models.PaymentServiceProvider{}, fmt.Errorf("%w", interfaces.ErrPaymentServiceProviderNotFound)
	}

	providers, err := r.loadPaymentServiceProviders(ctx)
	if err != nil {
		return models.PaymentServiceProvider{}, err
	}

	for _, provider := range providers {
		if provider.ID == providerID {
			return clonePaymentServiceProvider(provider), nil
		}
	}

	return models.PaymentServiceProvider{}, fmt.Errorf("%w", interfaces.ErrPaymentServiceProviderNotFound)
}

// ListInternationalMoneyTransferOperators returns the ordered list of current CBN-listed IMTO entries.
func (r *FinanceFileRepository) ListInternationalMoneyTransferOperators(ctx context.Context) ([]models.InternationalMoneyTransferOperator, error) {
	operators, err := r.loadInternationalMoneyTransferOperators(ctx)
	if err != nil {
		return nil, err
	}
	return cloneInternationalMoneyTransferOperatorList(operators), nil
}

// GetInternationalMoneyTransferOperator returns a single IMTO entry using its public slug identifier.
func (r *FinanceFileRepository) GetInternationalMoneyTransferOperator(ctx context.Context, operatorID string) (models.InternationalMoneyTransferOperator, error) {
	operatorID = strings.TrimSpace(operatorID)
	if operatorID == "" || !financeInternationalMoneyTransferOperatorIDPattern.MatchString(operatorID) {
		return models.InternationalMoneyTransferOperator{}, fmt.Errorf("%w", interfaces.ErrInternationalMoneyTransferOperatorNotFound)
	}

	operators, err := r.loadInternationalMoneyTransferOperators(ctx)
	if err != nil {
		return models.InternationalMoneyTransferOperator{}, err
	}

	for _, operator := range operators {
		if operator.ID == operatorID {
			return cloneInternationalMoneyTransferOperator(operator), nil
		}
	}

	return models.InternationalMoneyTransferOperator{}, fmt.Errorf("%w", interfaces.ErrInternationalMoneyTransferOperatorNotFound)
}

// ListCurrencies returns the ordered list of current monetary currencies.
func (r *FinanceFileRepository) ListCurrencies(ctx context.Context, filter interfaces.CurrencyFilter) ([]models.Currency, error) {
	currencies, countryIDs, err := r.loadCurrencies(ctx)
	if err != nil {
		return nil, err
	}
	countryAreaID := strings.TrimSpace(filter.CountryAreaID)
	if countryAreaID == "" {
		return cloneCurrencyList(currencies), nil
	}
	if !financeCurrencyCountryAreaIDPattern.MatchString(countryAreaID) {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidCurrencyCountryAreaID)
	}
	if _, ok := countryIDs[countryAreaID]; !ok {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidCurrencyCountryAreaID)
	}

	filtered := make([]models.Currency, 0)
	for _, currency := range currencies {
		for _, countryID := range currency.CountryAreaIDs {
			if countryID == countryAreaID {
				filtered = append(filtered, cloneCurrency(currency))
				break
			}
		}
	}
	return filtered, nil
}

// GetCurrency returns a single currency using its public alpha-3 code identifier.
func (r *FinanceFileRepository) GetCurrency(ctx context.Context, currencyID string) (models.Currency, error) {
	currencyID = strings.TrimSpace(currencyID)
	if currencyID == "" || !financeCurrencyIDPattern.MatchString(currencyID) {
		return models.Currency{}, fmt.Errorf("%w", interfaces.ErrCurrencyNotFound)
	}

	currencies, _, err := r.loadCurrencies(ctx)
	if err != nil {
		return models.Currency{}, err
	}

	for _, currency := range currencies {
		if currency.ID == currencyID {
			return cloneCurrency(currency), nil
		}
	}

	return models.Currency{}, fmt.Errorf("%w", interfaces.ErrCurrencyNotFound)
}

func (r *FinanceFileRepository) loadPaymentServiceProviders(ctx context.Context) ([]models.PaymentServiceProvider, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var providers []models.PaymentServiceProvider
	if err := r.jsonRepository.Decode(ctx, r.paymentServiceProvidersPath, &providers); err != nil {
		return nil, translateFinanceLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if providers == nil || len(providers) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := validatePaymentServiceProviders(providers); err != nil {
		return nil, err
	}

	return providers, nil
}

func (r *FinanceFileRepository) loadInternationalMoneyTransferOperators(ctx context.Context) ([]models.InternationalMoneyTransferOperator, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if strings.TrimSpace(r.internationalMoneyTransferOperatorsPath) == "" {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var operators []models.InternationalMoneyTransferOperator
	if err := r.jsonRepository.Decode(ctx, r.internationalMoneyTransferOperatorsPath, &operators); err != nil {
		return nil, translateFinanceLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if operators == nil || len(operators) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := validateInternationalMoneyTransferOperators(operators); err != nil {
		return nil, err
	}

	return operators, nil
}

func (r *FinanceFileRepository) loadCurrencies(ctx context.Context) ([]models.Currency, map[string]struct{}, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if strings.TrimSpace(r.currenciesPath) == "" || strings.TrimSpace(r.countriesAndAreasPath) == "" {
		return nil, nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	var currencies []models.Currency
	if err := r.jsonRepository.Decode(ctx, r.currenciesPath, &currencies); err != nil {
		return nil, nil, translateFinanceLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if currencies == nil || len(currencies) == 0 {
		return nil, nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	countryIDs, err := r.loadCountryAndAreaIDs(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := validateCurrencies(currencies, countryIDs); err != nil {
		return nil, nil, err
	}

	return currencies, countryIDs, nil
}

func (r *FinanceFileRepository) loadCountryAndAreaIDs(ctx context.Context) (map[string]struct{}, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var countries []models.CountryOrArea
	if err := r.jsonRepository.Decode(ctx, r.countriesAndAreasPath, &countries); err != nil {
		return nil, translateFinanceLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if countries == nil || len(countries) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if len(countries) != 248 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	allowed := make(map[string]struct{}, len(countries))
	for _, country := range countries {
		if country.ID == "" || !financeCurrencyCountryAreaIDPattern.MatchString(country.ID) {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := allowed[country.ID]; ok {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		allowed[country.ID] = struct{}{}
	}

	return allowed, nil
}

func validatePaymentServiceProviders(providers []models.PaymentServiceProvider) error {
	if len(providers) != 255 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	seenIDs := make(map[string]struct{}, len(providers))
	seenPairs := make(map[string]struct{}, len(providers))
	seenTypes := make(map[string]struct{}, len(financeExpectedPaymentServiceProviderCounts))
	categoryCounts := make(map[string]int, len(financeExpectedPaymentServiceProviderCounts))
	currentType := ""
	currentTypeOrder := -1
	lastName := ""
	lastID := ""

	for _, provider := range providers {
		if provider.ID == "" || provider.Name == "" || provider.InstitutionType == "" || provider.CountryCode == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if provider.CountryCode != "NG" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		typeOrder, ok := financePaymentServiceProviderTypeOrder[provider.InstitutionType]
		if !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !financePaymentServiceProviderIDPattern.MatchString(provider.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		typePrefix := strings.ReplaceAll(provider.InstitutionType, "_", "-")
		if !strings.HasPrefix(provider.ID, typePrefix+"-") {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if wantID := typePrefix + "-" + slugifyFinancePaymentServiceProviderName(provider.Name); provider.ID != wantID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[provider.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		pairKey := provider.InstitutionType + "\x00" + provider.Name
		if _, ok := seenPairs[pairKey]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		if currentType == "" {
			currentType = provider.InstitutionType
			currentTypeOrder = typeOrder
			lastName = ""
			lastID = ""
		} else if provider.InstitutionType != currentType {
			if typeOrder < currentTypeOrder {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, ok := seenTypes[provider.InstitutionType]; ok {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			seenTypes[currentType] = struct{}{}
			currentType = provider.InstitutionType
			currentTypeOrder = typeOrder
			lastName = ""
			lastID = ""
		}

		if currentTypeOrder != typeOrder {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if lastName != "" {
			if strings.Compare(strings.ToLower(lastName), strings.ToLower(provider.Name)) > 0 || (strings.Compare(strings.ToLower(lastName), strings.ToLower(provider.Name)) == 0 && strings.Compare(lastID, provider.ID) > 0) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}

		seenIDs[provider.ID] = struct{}{}
		seenPairs[pairKey] = struct{}{}
		categoryCounts[provider.InstitutionType]++
		lastName = provider.Name
		lastID = provider.ID
	}

	if currentType != "" {
		seenTypes[currentType] = struct{}{}
	}
	if len(seenTypes) != len(financeExpectedPaymentServiceProviderCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if !reflect.DeepEqual(categoryCounts, financeExpectedPaymentServiceProviderCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	return nil
}

func validateInternationalMoneyTransferOperators(operators []models.InternationalMoneyTransferOperator) error {
	if len(operators) != financeExpectedInternationalMoneyTransferOperatorCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	seenIDs := make(map[string]struct{}, len(operators))
	seenNames := make(map[string]struct{}, len(operators))
	prevName := ""
	prevID := ""

	for _, operator := range operators {
		if operator.ID == "" || operator.Name == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !financeInternationalMoneyTransferOperatorIDPattern.MatchString(operator.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if wantID := slugifyFinanceInternationalMoneyTransferOperatorName(operator.Name); operator.ID != wantID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if operator.Name == "OLIVE MONIES EXPRESS LIMITEDNOUVEAU MOBILE LIMITED" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := financeInternationalMoneyTransferOperatorFormerNames[strings.TrimSpace(operator.Name)]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		canonicalName := normalizeFinanceInternationalMoneyTransferOperatorName(operator.Name)
		if canonicalName == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[operator.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenNames[canonicalName]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		if prevName != "" {
			prevLower := strings.ToLower(prevName)
			currLower := strings.ToLower(operator.Name)
			if prevLower > currLower || (prevLower == currLower && prevID > operator.ID) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}

		seenIDs[operator.ID] = struct{}{}
		seenNames[canonicalName] = struct{}{}
		prevName = operator.Name
		prevID = operator.ID
	}

	if _, ok := seenNames["NOUVEAU MOBILE LIMITED"]; !ok {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if _, ok := seenNames["OLIVE MONIES EXPRESS LIMITED"]; !ok {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	return nil
}

func validateCurrencies(currencies []models.Currency, countryIDs map[string]struct{}) error {
	if len(currencies) != financeExpectedCurrencyCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	seenIDs := make(map[string]struct{}, len(currencies))
	seenAlphabetic := make(map[string]struct{}, len(currencies))
	seenNumeric := make(map[string]struct{}, len(currencies))
	countryToCodes := make(map[string][]string, financeExpectedCurrencyCountryAreaCount)
	relationships := 0
	zeroCount := 0
	prevName := ""
	prevAlphabetic := ""

	for _, currency := range currencies {
		if currency.ID == "" || currency.Name == "" || currency.AlphabeticCode == "" || currency.NumericCode == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !financeCurrencyIDPattern.MatchString(currency.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if currency.ID != strings.ToLower(currency.AlphabeticCode) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !financeCurrencyAlphabeticCodePattern.MatchString(currency.AlphabeticCode) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !financeCurrencyNumericCodePattern.MatchString(currency.NumericCode) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if currency.MinorUnit != 0 && currency.MinorUnit != 2 && currency.MinorUnit != 3 {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, forbidden := financeCurrencyExcludedCodes[currency.AlphabeticCode]; forbidden {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[currency.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenAlphabetic[currency.AlphabeticCode]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenNumeric[currency.NumericCode]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if currency.CountryAreaIDs == nil {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if len(currency.CountryAreaIDs) == 0 {
			zeroCount++
			if currency.AlphabeticCode != "TWD" {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if len(currency.CountryAreaIDs) > 0 {
			if !reflect.DeepEqual(currency.CountryAreaIDs, sortedFinanceStrings(currency.CountryAreaIDs)) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			seenCountries := make(map[string]struct{}, len(currency.CountryAreaIDs))
			for _, countryID := range currency.CountryAreaIDs {
				if !financeCurrencyCountryAreaIDPattern.MatchString(countryID) {
					return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
				}
				if _, ok := countryIDs[countryID]; !ok {
					return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
				}
				if _, ok := seenCountries[countryID]; ok {
					return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
				}
				seenCountries[countryID] = struct{}{}
				countryToCodes[countryID] = append(countryToCodes[countryID], currency.AlphabeticCode)
			}
		}

		if prevName != "" {
			if strings.Compare(prevName, currency.Name) > 0 || (strings.Compare(prevName, currency.Name) == 0 && strings.Compare(prevAlphabetic, currency.AlphabeticCode) > 0) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}

		seenIDs[currency.ID] = struct{}{}
		seenAlphabetic[currency.AlphabeticCode] = struct{}{}
		seenNumeric[currency.NumericCode] = struct{}{}
		relationships += len(currency.CountryAreaIDs)
		prevName = currency.Name
		prevAlphabetic = currency.AlphabeticCode
	}

	if zeroCount != 1 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if relationships != financeExpectedCurrencyRelationshipCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if len(countryToCodes) != financeExpectedCurrencyCountryAreaCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	for countryID := range countryIDs {
		_, isZeroMapping := financeCurrencyZeroMappingCountryAreaIDs[countryID]
		if isZeroMapping {
			if len(countryToCodes[countryID]) != 0 {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			continue
		}
		if len(countryToCodes[countryID]) == 0 {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}

	return nil
}

func clonePaymentServiceProviderList(providers []models.PaymentServiceProvider) []models.PaymentServiceProvider {
	if len(providers) == 0 {
		return make([]models.PaymentServiceProvider, 0)
	}
	cloned := make([]models.PaymentServiceProvider, len(providers))
	copy(cloned, providers)
	return cloned
}

func cloneCommercialBankList(banks []models.CommercialBank) []models.CommercialBank {
	if len(banks) == 0 {
		return make([]models.CommercialBank, 0)
	}
	cloned := make([]models.CommercialBank, len(banks))
	copy(cloned, banks)
	return cloned
}

func clonePaymentServiceProvider(provider models.PaymentServiceProvider) models.PaymentServiceProvider {
	return provider
}

func cloneInternationalMoneyTransferOperatorList(operators []models.InternationalMoneyTransferOperator) []models.InternationalMoneyTransferOperator {
	if len(operators) == 0 {
		return make([]models.InternationalMoneyTransferOperator, 0)
	}
	cloned := make([]models.InternationalMoneyTransferOperator, len(operators))
	copy(cloned, operators)
	return cloned
}

func cloneInternationalMoneyTransferOperator(operator models.InternationalMoneyTransferOperator) models.InternationalMoneyTransferOperator {
	return operator
}

func cloneCurrencyList(currencies []models.Currency) []models.Currency {
	if len(currencies) == 0 {
		return make([]models.Currency, 0)
	}
	cloned := make([]models.Currency, len(currencies))
	for i := range currencies {
		cloned[i] = cloneCurrency(currencies[i])
	}
	return cloned
}

func cloneCurrency(currency models.Currency) models.Currency {
	if len(currency.CountryAreaIDs) > 0 {
		currency.CountryAreaIDs = append([]string(nil), currency.CountryAreaIDs...)
	} else if currency.CountryAreaIDs == nil {
		currency.CountryAreaIDs = []string{}
	}
	return currency
}

func sortedFinanceStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	clone := append([]string(nil), values...)
	sort.Strings(clone)
	return clone
}

func validateFinanceDatasetPath(datasetPath string) (string, error) {
	datasetPath = strings.TrimSpace(datasetPath)
	if datasetPath == "" {
		return "", fmt.Errorf("payment service providers path is required")
	}
	if filepath.IsAbs(datasetPath) || filepath.VolumeName(datasetPath) != "" {
		return "", fmt.Errorf("payment service providers path must remain relative")
	}
	cleanedPath := filepath.Clean(datasetPath)
	if cleanedPath == "." || cleanedPath == ".." || strings.HasPrefix(cleanedPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("payment service providers path must remain relative")
	}
	return cleanedPath, nil
}

func slugifyFinancePaymentServiceProviderName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = financePaymentServiceProviderSlugPattern.ReplaceAllString(name, "-")
	name = financePaymentServiceProviderCollapsePattern.ReplaceAllString(name, "-")
	return strings.Trim(name, "-")
}

func slugifyFinanceInternationalMoneyTransferOperatorName(name string) string {
	name = normalizeFinanceInternationalMoneyTransferOperatorName(name)
	name = financeInternationalMoneyTransferOperatorSlugPattern.ReplaceAllString(strings.ToLower(name), "-")
	name = financeInternationalMoneyTransferOperatorCollapsePattern.ReplaceAllString(name, "-")
	return strings.Trim(name, "-")
}

func normalizeFinanceInternationalMoneyTransferOperatorName(name string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(name)), " ")
}

func translateFinanceLoadError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrDatasetFileNotFound):
		return fmt.Errorf("%w", interfaces.ErrDatasetFileNotFound)
	case errors.Is(err, interfaces.ErrDatasetFileTooLarge):
		return fmt.Errorf("%w", interfaces.ErrDatasetFileTooLarge)
	case errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	case errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	default:
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
}
