package assets

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type countryOrAreaFixture struct {
	ID string `json:"id"`
}

func TestEmbeddedFlagsOutsideWorkingDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	data, err := FlagSVG("ng")
	if err != nil {
		t.Fatalf("FlagSVG(ng) error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("FlagSVG(ng) returned empty data")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "flags", "4x3", "ng.svg")); !os.IsNotExist(err) {
		t.Fatalf("unexpected filesystem lookup success: %v", err)
	}
	if !bytes.HasPrefix(bytes.TrimSpace(data), []byte("<svg")) && !bytes.HasPrefix(bytes.TrimSpace(data), []byte("<?xml")) {
		t.Fatalf("embedded flag does not look like svg/xml: %q", string(bytes.TrimSpace(data)[:min(len(bytes.TrimSpace(data)), 40)]))
	}
	var doc struct {
		XMLName xml.Name `xml:"svg"`
		ViewBox string   `xml:"viewBox,attr"`
	}
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("embedded flag is not valid xml: %v", err)
	}
	if doc.ViewBox == "" {
		t.Fatal("embedded flag is missing viewBox")
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	datasetPath := filepath.Join(repoRoot, "datasets", "geography", "countries_and_areas.json")
	raw, err := os.ReadFile(datasetPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", datasetPath, err)
	}
	var countries []countryOrAreaFixture
	if err := json.Unmarshal(raw, &countries); err != nil {
		t.Fatalf("unmarshal countries: %v", err)
	}
	wantIDs := make([]string, 0, len(countries))
	for _, country := range countries {
		wantIDs = append(wantIDs, country.ID)
	}
	sort.Strings(wantIDs)

	entries, err := flagsFS.ReadDir("flags/4x3")
	if err != nil {
		t.Fatalf("ReadDir(flags/4x3) error = %v", err)
	}
	if len(entries) != 248 {
		t.Fatalf("unexpected embedded flag count: %d", len(entries))
	}
	gotIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			t.Fatalf("unexpected directory in embedded flag set: %s", name)
		}
		if filepath.Ext(name) != ".svg" || len(strings.TrimSuffix(name, ".svg")) != 2 {
			t.Fatalf("unexpected embedded flag name: %s", name)
		}
		gotIDs = append(gotIDs, strings.TrimSuffix(name, ".svg"))
	}
	sort.Strings(gotIDs)
	if len(gotIDs) != len(wantIDs) {
		t.Fatalf("unexpected embedded id count: got %d want %d", len(gotIDs), len(wantIDs))
	}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Fatalf("embedded flag mismatch at %d: got %q want %q", i, gotIDs[i], wantIDs[i])
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
