package assets

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed flags/4x3/*.svg
var flagsFS embed.FS

// FlagSVG returns the vendored 4x3 SVG flag asset for the provided lowercase alpha-2 code.
func FlagSVG(countryID string) ([]byte, error) {
	countryID = strings.TrimSpace(strings.ToLower(countryID))
	if len(countryID) != 2 {
		return nil, fs.ErrNotExist
	}
	return flagsFS.ReadFile("flags/4x3/" + countryID + ".svg")
}
