package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestFinanceRepositoryMicrofinanceBanksUsesRealDataset(t *testing.T) {
	root, err := filepath.Abs("../../../datasets")
	if err != nil {
		t.Fatal(err)
	}
	jsonRepository, err := NewJSONRepository(root, 16<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewFinanceRepository(jsonRepository, "finance/payment_service_providers.json", "finance/international_money_transfer_operators.json", "finance/non_interest_institutions.json", "finance/merchant_banks.json", "finance/payment_service_banks.json", "finance/financial_holding_companies.json", "finance/development_finance_institutions.json", "finance/primary_mortgage_institutions.json", financeMicrofinanceBanksRelativePath)
	if err != nil {
		t.Fatal(err)
	}
	banks, err := repository.ListMicrofinanceBanks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(banks) != 790 {
		t.Fatalf("microfinance bank count = %d, want 790", len(banks))
	}
	for _, name := range []string{"BWAY Microfinance bank Limited", "Katsu Microfinance Bank Limited", "Paragon Microfinance Bank Limited", "Teerus Microfinance Bank Limited", "Oganiru Microfinance Bank Limited"} {
		if !containsMicrofinanceBankName(banks, name) {
			t.Fatalf("retained anchor missing: %q", name)
		}
	}
	for _, name := range []string{"Bridgeway Microfinance Bank Limited", "Zigate Microfinance Bank Limited", "AKPO MICROFINANCE BANK LIMITED", "Verdant-Capital Microfinance Bank Limited"} {
		if containsMicrofinanceBankName(banks, name) {
			t.Fatalf("excluded anchor present: %q", name)
		}
	}
	found, err := repository.GetMicrofinanceBank(context.Background(), "bway-microfinance-bank-limited")
	if err != nil || found.Name != "BWAY Microfinance bank Limited" {
		t.Fatalf("GetMicrofinanceBank() = %#v, %v", found, err)
	}
	if _, err := repository.GetMicrofinanceBank(context.Background(), "missing-bank"); !errors.Is(err, interfaces.ErrMicrofinanceBankNotFound) {
		t.Fatalf("missing lookup error = %v", err)
	}
}

func TestFinanceRepositoryMicrofinanceBanksValidation(t *testing.T) {
	valid := loadMicrofinanceBanksDataset(t)
	tests := []struct {
		name   string
		mutate func([]models.MicrofinanceBank) []models.MicrofinanceBank
	}{
		{"incorrect record count", func(rows []models.MicrofinanceBank) []models.MicrofinanceBank { return rows[:789] }},
		{"duplicate ID", func(rows []models.MicrofinanceBank) []models.MicrofinanceBank { rows[1].ID = rows[0].ID; return rows }},
		{"duplicate name", func(rows []models.MicrofinanceBank) []models.MicrofinanceBank {
			rows[1].Name = rows[0].Name
			return rows
		}},
		{"invalid ID", func(rows []models.MicrofinanceBank) []models.MicrofinanceBank { rows[0].ID = "../invalid"; return rows }},
		{"wrong country", func(rows []models.MicrofinanceBank) []models.MicrofinanceBank {
			rows[0].CountryCode = "GH"
			return rows
		}},
		{"incorrect ordering", func(rows []models.MicrofinanceBank) []models.MicrofinanceBank {
			rows[0], rows[1] = rows[1], rows[0]
			return rows
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := append([]models.MicrofinanceBank(nil), valid...)
			rows = test.mutate(rows)
			repository := newMicrofinanceRepositoryForTest(t, rows, nil)
			if _, err := repository.ListMicrofinanceBanks(context.Background()); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("validation error = %v, want ErrInvalidDatasetFile", err)
			}
		})
	}
}

func TestFinanceRepositoryMicrofinanceBanksContextErrorsOwnershipAndCache(t *testing.T) {
	valid := loadMicrofinanceBanksDataset(t)
	stub := &microfinanceJSONRepository{raw: marshalMicrofinanceRows(t, valid)}
	repository, err := NewFinanceRepository(stub, "finance/payment_service_providers.json", "finance/international_money_transfer_operators.json", "finance/non_interest_institutions.json", "finance/merchant_banks.json", "finance/payment_service_banks.json", "finance/financial_holding_companies.json", "finance/development_finance_institutions.json", "finance/primary_mortgage_institutions.json", financeMicrofinanceBanksRelativePath)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repository.ListMicrofinanceBanks(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
	deadlineCtx, cancelDeadline := context.WithTimeout(context.Background(), time.Nanosecond)
	time.Sleep(time.Millisecond)
	defer cancelDeadline()
	if _, err := repository.ListMicrofinanceBanks(deadlineCtx); !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("deadline error = %v", err)
	}
	banks, err := repository.ListMicrofinanceBanks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	banks[0].Name = "changed"
	again, err := repository.ListMicrofinanceBanks(context.Background())
	if err != nil || again[0].Name == "changed" {
		t.Fatalf("cached slice was exposed: %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("decode calls = %d, want 1", stub.calls)
	}

	var group sync.WaitGroup
	for i := 0; i < 10; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := repository.ListMicrofinanceBanks(context.Background()); err != nil {
				t.Errorf("concurrent list error = %v", err)
			}
		}()
	}
	group.Wait()
}

func TestFinanceRepositoryMicrofinanceBanksSanitizesErrors(t *testing.T) {
	secret := errors.New("/private/tmp/microfinance_banks.json: permission denied")
	repository := newMicrofinanceRepositoryForTest(t, nil, secret)
	_, err := repository.ListMicrofinanceBanks(context.Background())
	if err == nil || strings.Contains(err.Error(), "/private/tmp") || !errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		t.Fatalf("unexpected sanitized error: %v", err)
	}
}

func TestFinanceRepositoryMicrofinanceBanksRejectsMalformedAndUnknownJSON(t *testing.T) {
	valid := marshalMicrofinanceRows(t, loadMicrofinanceBanksDataset(t))
	malformed := append([]json.RawMessage(nil), valid...)
	malformed[0] = json.RawMessage("{")
	unknown := append([]json.RawMessage(nil), valid...)
	unknown[0] = json.RawMessage(`{"id":"example-bank","name":"Example Bank","country_code":"NG","state":"Lagos"}`)
	for name, raw := range map[string][]json.RawMessage{
		"malformed JSON":   malformed,
		"unexpected field": unknown,
	} {
		t.Run(name, func(t *testing.T) {
			repository := newMicrofinanceRepositoryFromRaw(t, raw, nil)
			if _, err := repository.ListMicrofinanceBanks(context.Background()); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("validation error = %v, want ErrInvalidDatasetFile", err)
			}
		})
	}
}

type microfinanceJSONRepository struct {
	raw   []json.RawMessage
	err   error
	calls int
	mu    sync.Mutex
}

func (r *microfinanceJSONRepository) Decode(ctx context.Context, path string, destination any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	r.calls++
	r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	rows, ok := destination.(*[]json.RawMessage)
	if !ok {
		return fmt.Errorf("unexpected destination %T", destination)
	}
	*rows = append([]json.RawMessage(nil), r.raw...)
	return nil
}

func newMicrofinanceRepositoryForTest(t *testing.T, rows []models.MicrofinanceBank, decodeErr error) *FinanceFileRepository {
	t.Helper()
	return newMicrofinanceRepositoryFromRaw(t, marshalMicrofinanceRows(t, rows), decodeErr)
}

func newMicrofinanceRepositoryFromRaw(t *testing.T, raw []json.RawMessage, decodeErr error) *FinanceFileRepository {
	t.Helper()
	stub := &microfinanceJSONRepository{raw: raw, err: decodeErr}
	repository, err := NewFinanceRepository(stub, "finance/payment_service_providers.json", "finance/international_money_transfer_operators.json", "finance/non_interest_institutions.json", "finance/merchant_banks.json", "finance/payment_service_banks.json", "finance/financial_holding_companies.json", "finance/development_finance_institutions.json", "finance/primary_mortgage_institutions.json", financeMicrofinanceBanksRelativePath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

func loadMicrofinanceBanksDataset(t *testing.T) []models.MicrofinanceBank {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "datasets", "finance", "microfinance_banks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []models.MicrofinanceBank
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

func marshalMicrofinanceRows(t *testing.T, rows []models.MicrofinanceBank) []json.RawMessage {
	t.Helper()
	result := make([]json.RawMessage, len(rows))
	for i, row := range rows {
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		result[i] = encoded
	}
	return result
}

func containsMicrofinanceBankName(rows []models.MicrofinanceBank, name string) bool {
	for _, row := range rows {
		if row.Name == name {
			return true
		}
	}
	return false
}
