package assets

import (
	"embed"
	"io/fs"
	"regexp"
	"strings"
)

//go:embed financial-institutions/ng/*/*.png
var financialInstitutionLogosFS embed.FS

var financialInstitutionAssetPartPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// FinancialInstitutionLogo returns an embedded category-scoped logo.
func FinancialInstitutionLogo(category, institutionID, extension string) ([]byte, error) {
	category = strings.TrimSpace(strings.ToLower(category))
	institutionID = strings.TrimSpace(institutionID)
	extension = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(extension)), ".")
	if !financialInstitutionAssetPartPattern.MatchString(category) || !financialInstitutionAssetPartPattern.MatchString(institutionID) || extension != "png" {
		return nil, fs.ErrNotExist
	}
	return financialInstitutionLogosFS.ReadFile("financial-institutions/ng/" + category + "/" + institutionID + "." + extension)
}
