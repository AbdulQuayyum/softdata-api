package file

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var financeMicrofinanceBankIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var financeMicrofinanceBankCBNPattern = regexp.MustCompile(`^[0-9]{3}$`)
var financeMicrofinanceBankNIPPattern = regexp.MustCompile(`^[0-9]{6}$`)

var financeMicrofinanceBankRequiredNames = map[string]struct{}{
	"BWAY Microfinance bank Limited":    {},
	"Katsu Microfinance Bank Limited":   {},
	"Paragon Microfinance Bank Limited": {},
	"Teerus Microfinance Bank Limited":  {},
	"Oganiru Microfinance Bank Limited": {},
}

// ListMicrofinanceBanks returns the validated, ordered MFB roster.
func (r *FinanceFileRepository) ListMicrofinanceBanks(ctx context.Context) ([]models.MicrofinanceBank, error) {
	banks, err := r.loadMicrofinanceBanks(ctx)
	if err != nil {
		return nil, err
	}
	return cloneMicrofinanceBanks(banks), nil
}

// GetMicrofinanceBank returns one MFB using its exact public ID.
func (r *FinanceFileRepository) GetMicrofinanceBank(ctx context.Context, id string) (models.MicrofinanceBank, error) {
	id = strings.TrimSpace(id)
	if id == "" || !financeMicrofinanceBankIDPattern.MatchString(id) {
		return models.MicrofinanceBank{}, fmt.Errorf("%w", interfaces.ErrMicrofinanceBankNotFound)
	}
	banks, err := r.loadMicrofinanceBanks(ctx)
	if err != nil {
		return models.MicrofinanceBank{}, err
	}
	for _, bank := range banks {
		if bank.ID == id {
			return bank, nil
		}
	}
	return models.MicrofinanceBank{}, fmt.Errorf("%w", interfaces.ErrMicrofinanceBankNotFound)
}

func (r *FinanceFileRepository) loadMicrofinanceBanks(ctx context.Context) ([]models.MicrofinanceBank, error) {
	if r == nil || r.jsonRepository == nil || strings.TrimSpace(r.microfinanceBanksPath) == "" {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	const cacheKey = "finance:microfinance-banks"
	r.regulatedMu.RLock()
	if cached, ok := r.regulatedCache[cacheKey].([]models.MicrofinanceBank); ok {
		r.regulatedMu.RUnlock()
		return cloneMicrofinanceBanks(cached), nil
	}
	r.regulatedMu.RUnlock()

	var raw []json.RawMessage
	if err := r.jsonRepository.Decode(ctx, r.microfinanceBanksPath, &raw); err != nil {
		return nil, translateFinanceLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(raw) != 790 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	banks := make([]models.MicrofinanceBank, len(raw))
	for i, item := range raw {
		fields := map[string]json.RawMessage{}
		if err := json.Unmarshal(item, &fields); err != nil {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		for field := range fields {
			if field != "id" && field != "name" && field != "cbn_code" && field != "nip_code" && field != "website_url" && field != "logo_url" && field != "country_code" {
				return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if err := json.Unmarshal(item, &banks[i]); err != nil {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	if err := validateMicrofinanceBanks(banks); err != nil {
		return nil, err
	}

	r.regulatedMu.Lock()
	if existing, ok := r.regulatedCache[cacheKey].([]models.MicrofinanceBank); ok {
		banks = existing
	} else {
		r.regulatedCache[cacheKey] = banks
	}
	r.regulatedMu.Unlock()
	return cloneMicrofinanceBanks(banks), nil
}

func validateMicrofinanceBanks(banks []models.MicrofinanceBank) error {
	if len(banks) != 790 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	seenIDs := make(map[string]struct{}, len(banks))
	seenNames := make(map[string]struct{}, len(banks))
	seenCBN := make(map[string]struct{}, len(banks))
	seenNIP := make(map[string]struct{}, len(banks))
	previousName, previousID := "", ""
	for _, bank := range banks {
		if bank.ID == "" || bank.Name == "" || strings.TrimSpace(bank.Name) != bank.Name || bank.CountryCode != "NG" || !financeMicrofinanceBankIDPattern.MatchString(bank.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, exists := seenIDs[bank.ID]; exists {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, exists := seenNames[bank.Name]; exists {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if previousName != "" && (strings.ToLower(previousName) > strings.ToLower(bank.Name) || (strings.EqualFold(previousName, bank.Name) && previousID > bank.ID)) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[bank.ID] = struct{}{}
		seenNames[bank.Name] = struct{}{}
		if bank.CBNCode != "" {
			if !financeMicrofinanceBankCBNPattern.MatchString(bank.CBNCode) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, exists := seenCBN[bank.CBNCode]; exists {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			seenCBN[bank.CBNCode] = struct{}{}
		}
		if bank.NIPCode != "" {
			if !financeMicrofinanceBankNIPPattern.MatchString(bank.NIPCode) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, exists := seenNIP[bank.NIPCode]; exists {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			seenNIP[bank.NIPCode] = struct{}{}
		}
		previousName, previousID = bank.Name, bank.ID
	}
	for name := range financeMicrofinanceBankRequiredNames {
		if _, exists := seenNames[name]; !exists {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	for _, name := range []string{"AKPO MICROFINANCE BANK LIMITED", "Verdant-Capital Microfinance Bank Limited", "Bridgeway Microfinance Bank Limited", "Zigate Microfinance Bank Limited"} {
		if _, exists := seenNames[name]; exists {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	return nil
}

func cloneMicrofinanceBanks(banks []models.MicrofinanceBank) []models.MicrofinanceBank {
	if len(banks) == 0 {
		return make([]models.MicrofinanceBank, 0)
	}
	cloned := make([]models.MicrofinanceBank, len(banks))
	copy(cloned, banks)
	return cloned
}
