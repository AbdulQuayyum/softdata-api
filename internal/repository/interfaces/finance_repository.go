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
}
