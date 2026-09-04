package validators

import "testing"

func TestValidateCommercialBankID(t *testing.T) {
	valid := []string{"access-bank", "alpha-morgan-bank", "citibank-nigeria", "ecobank-nigeria", "fidelity-bank", "first-bank-of-nigeria", "first-city-monument-bank", "globus-bank", "guaranty-trust-bank", "keystone-bank", "nova-bank", "optimus-bank", "parallex-bank", "polaris-bank", "premium-trust-bank", "providus-bank", "signature-bank", "stanbic-ibtc-bank", "standard-chartered-bank", "sterling-bank", "suntrust-bank", "tatum-bank", "titan-trust-bank", "union-bank", "united-bank-for-africa", "unity-bank", "wema-bank", "zenith-bank"}
	for _, value := range valid {
		if err := ValidateCommercialBankID(value); err != nil {
			t.Errorf("ValidateCommercialBankID(%q) error = %v", value, err)
		}
	}
	for _, value := range []string{"", " ", "Access-bank", "unknown-bank", "-access-bank", "access-bank-", "access--bank", "access bank", "access/bank", `access\\bank`, ".", "../access-bank", "access%2Dbank", "access-bank?x=1", "access-bank#x"} {
		if err := ValidateCommercialBankID(value); err == nil {
			t.Errorf("ValidateCommercialBankID(%q) unexpectedly succeeded", value)
		}
	}
}
