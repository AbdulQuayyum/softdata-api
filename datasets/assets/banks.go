package assets

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed banks/ng/*.png
var bankLogosFS embed.FS

// BankLogo returns the embedded logo for a commercial-bank ID and extension.
func BankLogo(bankID, extension string) ([]byte, error) {
	bankID = strings.TrimSpace(bankID)
	extension = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(extension)), ".")
	if bankID == "" || extension != "png" || strings.ContainsAny(bankID, "/\\") {
		return nil, fs.ErrNotExist
	}
	return bankLogosFS.ReadFile("banks/ng/" + bankID + "." + extension)
}
