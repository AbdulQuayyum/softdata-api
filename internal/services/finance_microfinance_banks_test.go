package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestFinanceServiceMicrofinanceBankListAndLookup(t *testing.T) {
	bank := models.MicrofinanceBank{ID: "bway-microfinance-bank-limited", Name: "BWAY Microfinance bank Limited", CountryCode: "NG"}
	service, err := NewFinanceService(&financeRepositoryStub{
		microResult: []models.MicrofinanceBank{bank},
		microByID:   map[string]models.MicrofinanceBank{bank.ID: bank},
	})
	if err != nil {
		t.Fatal(err)
	}
	list, err := service.ListMicrofinanceBanks(context.Background())
	if err != nil || len(list) != 1 || list[0] != bank {
		t.Fatalf("ListMicrofinanceBanks() = %#v, %v", list, err)
	}
	list[0].Name = "changed"
	found, err := service.GetMicrofinanceBank(context.Background(), " bway-microfinance-bank-limited ")
	if err != nil || found != bank {
		t.Fatalf("GetMicrofinanceBank() = %#v, %v", found, err)
	}
}

func TestFinanceServiceMicrofinanceBankValidationAndErrors(t *testing.T) {
	stub := &financeRepositoryStub{
		microGetErr: errors.New("secret /tmp/datasets/microfinance_banks.json failure"),
	}
	service, err := NewFinanceService(stub)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"", "bad id", "../bank", "bank/id", "bank?x=1", "bank#fragment", "bank%2Did"} {
		if _, err := service.GetMicrofinanceBank(context.Background(), id); !errors.Is(err, ErrInvalidMicrofinanceBankID) {
			t.Fatalf("GetMicrofinanceBank(%q) error = %v, want invalid ID", id, err)
		}
	}
	if stub.microGetCalls != 0 {
		t.Fatal("invalid IDs reached the repository")
	}
	stub.microGetErr = nil
	if _, err := service.GetMicrofinanceBank(context.Background(), "missing-bank"); !errors.Is(err, ErrMicrofinanceBankNotFound) {
		t.Fatalf("missing lookup error = %v", err)
	}
	stub.microGetErr = errors.New("secret /tmp/datasets/microfinance_banks.json failure")
	if _, err := service.GetMicrofinanceBank(context.Background(), "secret-bank"); err == nil || !strings.Contains(err.Error(), "repository unavailable") || strings.Contains(err.Error(), "/tmp/datasets") {
		t.Fatalf("unexpected sanitized error: %v", err)
	}
}

func TestFinanceServiceMicrofinanceBankContextPropagation(t *testing.T) {
	service, err := NewFinanceService(&financeRepositoryStub{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.ListMicrofinanceBanks(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("list cancellation error = %v", err)
	}
	deadlineCtx, cancelDeadline := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancelDeadline()
	time.Sleep(time.Millisecond)
	if _, err := service.GetMicrofinanceBank(deadlineCtx, "valid-bank"); !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("lookup deadline error = %v", err)
	}

	service, err = NewFinanceService(&financeRepositoryStub{microGetErr: interfaces.ErrMicrofinanceBankNotFound})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetMicrofinanceBank(context.Background(), "missing-bank"); !errors.Is(err, ErrMicrofinanceBankNotFound) {
		t.Fatalf("translated not-found error = %v", err)
	}
}
