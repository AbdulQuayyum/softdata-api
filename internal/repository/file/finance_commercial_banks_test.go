package file

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestFinanceRepositoryCommercialBanksListLookupAndOwnership(t *testing.T) {
	root := t.TempDir()
	data := readFinanceDatasetBytes(t, financeDatasetPath("finance", "commercial_banks.json"))
	if err := os.MkdirAll(filepath.Join(root, "finance"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "finance", "commercial_banks.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	jsonRepo, err := NewJSONRepository(root, int64(len(data)+1))
	if err != nil {
		t.Fatal(err)
	}
	repo, err := NewFinanceRepository(jsonRepo, "finance/payment_service_providers.json")
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.ListCommercialBanks(context.Background())
	if err != nil {
		t.Fatalf("ListCommercialBanks() error = %v", err)
	}
	if len(got) != 28 || got == nil {
		t.Fatalf("ListCommercialBanks() = %d records, want non-nil 28", len(got))
	}
	if got[0].CBNCode != "044" || got[0].NIPCode != "000014" {
		t.Fatalf("unexpected Access Bank identifiers: %#v", got[0])
	}
	nova, err := repo.GetCommercialBank(context.Background(), " nova-bank ")
	if err != nil {
		t.Fatal(err)
	}
	if nova.CBNCode != "" || nova.NIPCode != "" {
		t.Fatalf("NOVA identifiers = (%q, %q), want omitted", nova.CBNCode, nova.NIPCode)
	}
	got[0].Name = "changed"
	again, err := repo.ListCommercialBanks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if again[0].Name == "changed" {
		t.Fatal("ListCommercialBanks() returned shared slice state")
	}
	if _, err := repo.GetCommercialBank(context.Background(), "missing-bank"); !errors.Is(err, interfaces.ErrCommercialBankNotFound) {
		t.Fatalf("unknown lookup error = %v", err)
	}
}

func TestFinanceRepositoryCommercialBanksPreservesContextAndMaxBytes(t *testing.T) {
	root := t.TempDir()
	data := readFinanceDatasetBytes(t, financeDatasetPath("finance", "commercial_banks.json"))
	if err := os.MkdirAll(filepath.Join(root, "finance"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "finance", "commercial_banks.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	jsonRepo, err := NewJSONRepository(root, int64(len(data)-1))
	if err != nil {
		t.Fatal(err)
	}
	repo, err := NewFinanceRepository(jsonRepo, "finance/payment_service_providers.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ListCommercialBanks(context.Background()); !errors.Is(err, interfaces.ErrDatasetFileTooLarge) {
		t.Fatalf("max-byte error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.ListCommercialBanks(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context error = %v", err)
	}
}

func TestCommercialBankCloneIsValueStable(t *testing.T) {
	input := []models.CommercialBank{{ID: "access-bank", Name: "Access Bank Plc"}}
	output := cloneCommercialBankList(input)
	if !reflect.DeepEqual(input, output) || &input[0] == &output[0] {
		t.Fatal("commercial bank clone did not create an independent slice")
	}
}
