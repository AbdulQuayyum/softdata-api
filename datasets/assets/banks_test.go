package assets

import (
	"bytes"
	"image/png"
	"testing"
)

var commercialBankAssetIDs = []string{
	"access-bank", "alpha-morgan-bank", "citibank-nigeria", "ecobank-nigeria", "fidelity-bank",
	"first-bank-of-nigeria", "first-city-monument-bank", "globus-bank", "guaranty-trust-bank", "keystone-bank",
	"nova-bank", "optimus-bank", "parallex-bank", "polaris-bank", "premium-trust-bank", "providus-bank",
	"signature-bank", "stanbic-ibtc-bank", "standard-chartered-bank", "sterling-bank", "suntrust-bank", "tatum-bank",
	"titan-trust-bank", "union-bank", "united-bank-for-africa", "unity-bank", "wema-bank", "zenith-bank",
}

func TestCommercialBankLogosAreCompletePNGAssets(t *testing.T) {
	if len(commercialBankAssetIDs) != 28 {
		t.Fatalf("asset ID count = %d, want 28", len(commercialBankAssetIDs))
	}
	for _, id := range commercialBankAssetIDs {
		data, err := BankLogo(id, "png")
		if err != nil {
			t.Fatalf("BankLogo(%q): %v", id, err)
		}
		if !bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) {
			t.Fatalf("BankLogo(%q) is not a PNG", id)
		}
		if _, err := png.Decode(bytes.NewReader(data)); err != nil {
			t.Fatalf("png.Decode(%q): %v", id, err)
		}
	}
}
