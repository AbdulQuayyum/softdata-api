package file

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestRegulatedFinanceDatasetsLoadTogether(t *testing.T) {
	root, err := filepath.Abs("../../../datasets")
	if err != nil {
		t.Fatal(err)
	}
	jsonRepository, err := NewJSONRepository(root, 16<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewFinanceRepository(jsonRepository,
		"finance/payment_service_providers.json",
		"finance/international_money_transfer_operators.json",
		"finance/non_interest_institutions.json",
		"finance/merchant_banks.json",
		"finance/payment_service_banks.json",
		"finance/financial_holding_companies.json",
		"finance/development_finance_institutions.json",
		"finance/primary_mortgage_institutions.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	counts := 0
	nonInterest, err := repository.ListNonInterestFinancialInstitutions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts += len(nonInterest)
	merchant, err := repository.ListMerchantBanks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts += len(merchant)
	paymentService, err := repository.ListPaymentServiceBanks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts += len(paymentService)
	holdings, err := repository.ListFinancialHoldingCompanies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts += len(holdings)
	development, err := repository.ListDevelopmentFinanceInstitutions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts += len(development)
	mortgages, err := repository.ListPrimaryMortgageInstitutions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts += len(mortgages)
	if counts != 63 {
		t.Fatalf("combined record count = %d, want 63", counts)
	}
	if len(mortgages) != 31 {
		t.Fatalf("mortgage record count = %d, want 31", len(mortgages))
	}
	if _, err := repository.GetPrimaryMortgageInstitution(context.Background(), "trustbond-mortgage-bank"); err == nil {
		t.Fatal("TrustBond should not be present")
	}
	if _, err := repository.GetPrimaryMortgageInstitution(context.Background(), "firsttrust-mortgage-bank"); err != nil {
		t.Fatalf("FirstTrust lookup failed: %v", err)
	}
	akwa, err := repository.GetPrimaryMortgageInstitution(context.Background(), "akwa-savings")
	if err != nil {
		t.Fatal(err)
	}
	if akwa.Name != "Ibom Mortgage Bank" || akwa.WebsiteURL != "https://ibommortgagebank.com/" || akwa.LogoURL == "" {
		t.Fatalf("Akwa/Ibom record = %#v", akwa)
	}
	var _ models.PrimaryMortgageInstitution = mortgages[0]
}
