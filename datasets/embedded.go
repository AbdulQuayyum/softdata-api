package datasets

import (
	"embed"
	"io/fs"
)

// FS contains the JSON documents required by the production API. Vercel's Go
// runtime does not automatically include sibling data directories beside a
// compiled function, so the serverless bootstrap can use this filesystem.
//
//go:embed education/*.json finance/*.json geography/*.json
var FS embed.FS

// Files returns the embedded runtime dataset filesystem.
func Files() fs.FS {
	return FS
}
