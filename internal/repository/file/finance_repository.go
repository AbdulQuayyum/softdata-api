package file

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var financePaymentServiceProviderIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
var financePaymentServiceProviderSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)
var financePaymentServiceProviderCollapsePattern = regexp.MustCompile(`-+`)

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

// FinanceFileRepository reads payment-service-provider records from a JSON dataset file.
type FinanceFileRepository struct {
	jsonRepository              interfaces.JSONFileRepository
	paymentServiceProvidersPath string
}

var _ interfaces.FinanceRepository = (*FinanceFileRepository)(nil)

// NewFinanceRepository constructs a file-backed finance repository.
func NewFinanceRepository(jsonRepository interfaces.JSONFileRepository, paymentServiceProvidersPath string) (*FinanceFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanedPath, err := validateFinanceDatasetPath(paymentServiceProvidersPath)
	if err != nil {
		return nil, err
	}

	return &FinanceFileRepository{
		jsonRepository:              jsonRepository,
		paymentServiceProvidersPath: cleanedPath,
	}, nil
}

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

func clonePaymentServiceProviderList(providers []models.PaymentServiceProvider) []models.PaymentServiceProvider {
	if len(providers) == 0 {
		return make([]models.PaymentServiceProvider, 0)
	}
	cloned := make([]models.PaymentServiceProvider, len(providers))
	copy(cloned, providers)
	return cloned
}

func clonePaymentServiceProvider(provider models.PaymentServiceProvider) models.PaymentServiceProvider {
	return provider
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

func translateFinanceLoadError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrDatasetFileNotFound):
		return fmt.Errorf("%w", interfaces.ErrDatasetFileNotFound)
	case errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	case errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	default:
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
}
