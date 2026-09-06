package validators

import "testing"

func TestValidateMicrofinanceBankID(t *testing.T) {
	for _, id := range []string{
		"bway-microfinance-bank-limited",
		"verdant",
		"bank9",
		"a",
	} {
		if err := ValidateMicrofinanceBankID(id); err != nil {
			t.Errorf("ValidateMicrofinanceBankID(%q) error = %v", id, err)
		}
	}

	for _, id := range []string{
		"",
		" Uppercase",
		"UPPERCASE",
		" valid-id ",
		"has space",
		"-leading",
		"trailing-",
		"double--hyphen",
		"../secret",
		"a/b",
		"a\\b",
		"a?b",
		"a#b",
		"a%b",
		"550e8400-e29b-41d4-a716-446655440000",
		string(make([]byte, financeMicrofinanceBankIDMaxLength+1)),
	} {
		if err := ValidateMicrofinanceBankID(id); err == nil {
			t.Errorf("ValidateMicrofinanceBankID(%q) error = nil, want error", id)
		}
	}
}
