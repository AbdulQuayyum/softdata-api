package validators

import (
	"net/url"
	"strings"
	"testing"
)

func TestAccreditationValidatorBoundaries(t *testing.T) {
	for _, n := range []int{100, 101} {
		q, err := ValidateMedicalLaboratoryAccreditationListQuery(url.Values{"search": {" " + strings.Repeat("é", n) + " "}})
		if (err == nil) != (n == 100) {
			t.Fatalf("search length %d: %v", n, err)
		}
		if err == nil && len([]rune(q.Search)) != 100 {
			t.Fatal("search normalization")
		}
	}
	q, err := ValidateMedicalLaboratoryAccreditationListQuery(url.Values{"state_id": {"unknown-state"}})
	if err != nil || q.StateID != "unknown-state" {
		t.Fatal("state existence must be delegated")
	}
	for _, n := range []int{255, 256} {
		_, err := ValidateMedicalLaboratoryAccreditationID("accreditation_id", strings.Repeat("a", n))
		if (err == nil) != (n == 255) {
			t.Fatalf("ID length %d: %v", n, err)
		}
	}
}
