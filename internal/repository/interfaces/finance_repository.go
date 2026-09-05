package interfaces

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// CurrencyFilter narrows currency list results by country or area.
type CurrencyFilter struct {
	CountryAreaID string
}

// FinanceRepository defines payment-service-provider lookup operations backed by a finance dataset.
type FinanceRepository interface {
	ListPaymentServiceProviders(ctx context.Context) ([]models.PaymentServiceProvider, error)
	ListPaymentServiceProvidersByType(ctx context.Context, institutionType string) ([]models.PaymentServiceProvider, error)
	GetPaymentServiceProvider(ctx context.Context, providerID string) (models.PaymentServiceProvider, error)
	ListInternationalMoneyTransferOperators(ctx context.Context) ([]models.InternationalMoneyTransferOperator, error)
	GetInternationalMoneyTransferOperator(ctx context.Context, operatorID string) (models.InternationalMoneyTransferOperator, error)
	ListCurrencies(ctx context.Context, filter CurrencyFilter) ([]models.Currency, error)
	GetCurrency(ctx context.Context, currencyID string) (models.Currency, error)
	ListCommercialBanks(ctx context.Context) ([]models.CommercialBank, error)
	GetCommercialBank(ctx context.Context, bankID string) (models.CommercialBank, error)
	ListNonInterestFinancialInstitutions(ctx context.Context) ([]models.NonInterestInstitution, error)
	GetNonInterestFinancialInstitution(ctx context.Context, id string) (models.NonInterestInstitution, error)
	ListMerchantBanks(ctx context.Context) ([]models.MerchantBank, error)
	GetMerchantBank(ctx context.Context, id string) (models.MerchantBank, error)
	ListPaymentServiceBanks(ctx context.Context) ([]models.PaymentServiceBank, error)
	GetPaymentServiceBank(ctx context.Context, id string) (models.PaymentServiceBank, error)
	ListFinancialHoldingCompanies(ctx context.Context) ([]models.FinancialHoldingCompany, error)
	GetFinancialHoldingCompany(ctx context.Context, id string) (models.FinancialHoldingCompany, error)
	ListDevelopmentFinanceInstitutions(ctx context.Context) ([]models.DevelopmentFinanceInstitution, error)
	GetDevelopmentFinanceInstitution(ctx context.Context, id string) (models.DevelopmentFinanceInstitution, error)
	ListPrimaryMortgageInstitutions(ctx context.Context) ([]models.PrimaryMortgageInstitution, error)
	GetPrimaryMortgageInstitution(ctx context.Context, id string) (models.PrimaryMortgageInstitution, error)
}
