package handlers

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

const (
	openAPIPath  = "openapi.yaml"
	postmanPath  = "softdata-api.postman_collection.json"
	docsAPITitle = "SoftData API"
)

var (
	landingTemplate = template.Must(template.New("landing").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>SoftData API</title>
  <style>
    :root { color-scheme: light dark; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    body { margin: 0; background: #f8fafc; color: #111827; }
    main { max-width: 880px; margin: 0 auto; padding: 64px 24px; }
    h1 { font-size: clamp(2rem, 6vw, 4rem); line-height: 1; margin: 0 0 18px; letter-spacing: 0; }
    p { font-size: 1.075rem; line-height: 1.7; max-width: 720px; color: #374151; }
    .links { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 32px; }
    a { color: #0f766e; font-weight: 650; }
    .button { border: 1px solid #cbd5e1; border-radius: 8px; padding: 10px 14px; text-decoration: none; background: #fff; color: #111827; }
    .primary { background: #111827; color: #fff; border-color: #111827; }
    code { background: #e5e7eb; border-radius: 6px; padding: 2px 6px; }
    @media (prefers-color-scheme: dark) {
      body { background: #020617; color: #f8fafc; }
      p { color: #cbd5e1; }
      .button { background: #0f172a; border-color: #334155; color: #f8fafc; }
      .primary { background: #f8fafc; color: #020617; border-color: #f8fafc; }
      code { background: #1e293b; }
    }
  </style>
</head>
<body>
  <main>
    <h1>SoftData API</h1>
    <p>Structured, normalized and developer-ready public datasets. Anonymous access is supported for public endpoints; optional <code>X-API-Key</code> authentication is available for higher limits and account-level usage analytics.</p>
    <p><code>docs/openapi.yaml</code> is the canonical API contract. The browser docs, downloadable OpenAPI file and Postman collection are synchronized from that contract.</p>
    <div class="links">
      <a class="button primary" href="/docs">API Documentation</a>
      <a class="button" href="/openapi.yaml">OpenAPI Specification</a>
      <a class="button" href="/postman.json">Postman Collection</a>
      <a class="button" href="https://github.com/AbdulQuayyum/softdata-api" rel="noopener noreferrer">GitHub</a>
    </div>
  </main>
</body>
</html>
`))

	docsTemplate = template.Must(template.New("docs").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>SoftData API Documentation</title>
  <style>
    body { margin: 0; }
    .topbar { align-items: center; background: #0f172a; color: #f8fafc; display: flex; flex-wrap: wrap; gap: 12px; font: 14px/1.5 Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; justify-content: space-between; padding: 10px 16px; }
    .topbar a { color: #99f6e4; text-decoration: none; }
  </style>
</head>
<body>
  <div class="topbar">
    <span>SoftData API documentation. Optional API keys can be entered in the Authorize panel and cleared at any time.</span>
    <span><a href="/openapi.yaml">OpenAPI</a> · <a href="/postman.json">Postman</a> · <a href="https://github.com/AbdulQuayyum/softdata-api" rel="noopener noreferrer">GitHub</a></span>
  </div>
  <script id="api-reference" data-url="/openapi.yaml"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.25.23"></script>
</body>
</html>
`))
)

// DocumentationHandler serves the human and machine-readable API docs.
type DocumentationHandler struct {
	fs fs.FS
}

// NewDocumentationHandler constructs a handler for committed documentation artifacts.
func NewDocumentationHandler(files fs.FS) (*DocumentationHandler, error) {
	if files == nil {
		return nil, fmt.Errorf("documentation filesystem is required")
	}
	if _, err := fs.ReadFile(files, openAPIPath); err != nil {
		return nil, err
	}
	if _, err := fs.ReadFile(files, postmanPath); err != nil {
		return nil, err
	}
	return &DocumentationHandler{fs: files}, nil
}

func (h *DocumentationHandler) ServeLanding(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, landingTemplate)
}

func (h *DocumentationHandler) ServeDocs(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, docsTemplate)
}

func (h *DocumentationHandler) ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	h.serveEmbedded(w, openAPIPath, "application/yaml; charset=utf-8")
}

func (h *DocumentationHandler) ServePostman(w http.ResponseWriter, r *http.Request) {
	h.serveEmbedded(w, postmanPath, "application/json; charset=utf-8")
}

func (h *DocumentationHandler) serveEmbedded(w http.ResponseWriter, path, contentType string) {
	if h == nil || h.fs == nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	body, err := fs.ReadFile(h.fs, path)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, _ = w.Write(body)
}

func writeHTML(w http.ResponseWriter, tmpl *template.Template) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if err := tmpl.Execute(w, nil); err != nil && !strings.Contains(err.Error(), "broken pipe") {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
