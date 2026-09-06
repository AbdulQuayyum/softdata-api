package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestVerifyMicrofinanceBankDatasetAcceptsManifestAndCallsServiceOnce(t *testing.T) {
	banks := loadMicrofinanceBanks(t)
	service := &financeServiceStub{microfinanceBanks: banks}

	if err := verifyMicrofinanceBankDataset(context.Background(), service); err != nil {
		t.Fatalf("verifyMicrofinanceBankDataset() error = %v", err)
	}
	if service.microfinanceCalls != 1 {
		t.Fatalf("ListMicrofinanceBanks() calls = %d, want 1", service.microfinanceCalls)
	}
}

func TestVerifyMicrofinanceBankDatasetRejectsManifestViolations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]models.MicrofinanceBank)
	}{
		{name: "wrong count", mutate: func(banks []models.MicrofinanceBank) { banks = banks[:len(banks)-1] }},
		{name: "duplicate id", mutate: func(banks []models.MicrofinanceBank) { banks[1].ID = banks[0].ID }},
		{name: "duplicate name", mutate: func(banks []models.MicrofinanceBank) { banks[1].Name = banks[0].Name }},
		{name: "invalid id", mutate: func(banks []models.MicrofinanceBank) { banks[0].ID = "BWAY" }},
		{name: "wrong country", mutate: func(banks []models.MicrofinanceBank) { banks[0].CountryCode = "GH" }},
		{name: "incorrect ordering", mutate: func(banks []models.MicrofinanceBank) { banks[0], banks[1] = banks[1], banks[0] }},
		{name: "missing required anchor", mutate: func(banks []models.MicrofinanceBank) {
			replaceMicrofinanceBankID(banks, "bway-microfinance-bank-limited", "replacement-bank")
		}},
		{name: "prohibited anchor", mutate: func(banks []models.MicrofinanceBank) {
			replaceMicrofinanceBankID(banks, "bway-microfinance-bank-limited", "bridgeway-microfinance-bank-limited")
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			banks := loadMicrofinanceBanks(t)
			if tc.name == "wrong count" {
				banks = banks[:len(banks)-1]
			} else {
				tc.mutate(banks)
			}
			if err := verifyMicrofinanceBankDataset(context.Background(), &financeServiceStub{microfinanceBanks: banks}); err == nil {
				t.Fatal("verifyMicrofinanceBankDataset() error = nil, want error")
			}
		})
	}
}

func TestVerifyMicrofinanceBankDatasetPreservesContextAndSanitizesFailures(t *testing.T) {
	banks := loadMicrofinanceBanks(t)
	for _, tc := range []struct {
		name string
		ctx  context.Context
		want error
	}{
		{name: "canceled", ctx: canceledContext(), want: context.Canceled},
		{name: "deadline", ctx: expiredContext(), want: context.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyMicrofinanceBankDataset(tc.ctx, &financeServiceStub{microfinanceBanks: banks})
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}

	serviceErr := errors.New("repository unavailable: /private/datasets/microfinance_banks.json")
	err := verifyMicrofinanceBankDataset(context.Background(), &financeServiceStub{microfinanceErr: serviceErr})
	if err == nil || !strings.Contains(err.Error(), "verify microfinance bank dataset") || !errors.Is(err, serviceErr) {
		t.Fatalf("unexpected service error: %v", err)
	}
}

func loadMicrofinanceBanks(t *testing.T) []models.MicrofinanceBank {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean("../../datasets/finance/microfinance_banks.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var banks []models.MicrofinanceBank
	if err := json.Unmarshal(data, &banks); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return banks
}

func replaceMicrofinanceBankID(banks []models.MicrofinanceBank, from, to string) {
	for i := range banks {
		if banks[i].ID == from {
			banks[i].ID = to
			return
		}
	}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func expiredContext() context.Context {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	return ctx
}
