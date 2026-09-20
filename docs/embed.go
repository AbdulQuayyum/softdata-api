package docs

import "embed"

// FS contains the canonical API documentation artifacts served by the API.
//
//go:embed openapi.yaml softdata-api.postman_collection.json
var FS embed.FS
