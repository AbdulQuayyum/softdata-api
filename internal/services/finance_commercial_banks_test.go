package services

import (
	"context"
	"errors"
	"testing"
)

func TestFinanceServiceCommercialBankLookupAndValidation(t *testing.T) {
	stub := &financeRepositoryStub{}
	svc, err := NewFinanceService(stub)
	if err != nil {
		t.Fatal(err)
	}
	bank, err := svc.GetCommercialBank(context.Background(), " access-bank ")
	if err != nil || bank.CBNCode != "044" || bank.NIPCode != "000014" {
		t.Fatalf("GetCommercialBank() = %#v, %v", bank, err)
	}
	for _, id := range []string{"", "BANK", "bad id", "../access-bank", "access/bank", "access%2Dbank", "access-bank?x=1"} {
		if _, err := svc.GetCommercialBank(context.Background(), id); !errors.Is(err, ErrInvalidCommercialBankID) {
			t.Fatalf("GetCommercialBank(%q) error = %v", id, err)
		}
	}
	if _, err := svc.GetCommercialBank(context.Background(), "missing-bank"); !errors.Is(err, ErrCommercialBankNotFound) {
		t.Fatalf("missing bank error = %v", err)
	}
	if _, err := svc.GetCommercialBank(context.Background(), " access-bank "); err != nil {
		t.Fatal(err)
	}
}
