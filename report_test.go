package konform

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReportSourcesAndPrint(t *testing.T) {
	type config struct {
		Server struct {
			Port int    `key:"server.port" default:"8080"`
			Host string `key:"server.host" default:"127.0.0.1"`
		}
		Database struct {
			URL string `key:"database.url" env:"DATABASE_URL" secret:"true"`
		}
		Log struct {
			Level string `key:"log.level"`
		}
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"server":{"port":"9090"},"log":{"level":"debug"}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/app")

	cfg := &config{}
	report, err := Load(cfg, FromJSONFile(path), FromEnv())
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if report == nil {
		t.Fatalf("Load() report = nil, want non-nil")
	}

	entries := make(map[string]ReportEntry, len(report.Entries))
	for _, entry := range report.Entries {
		entries[entry.Path] = entry
	}

	if got := entries["server.port"].Source; got != path {
		t.Fatalf(`report source for "server.port" = %q, want %q`, got, path)
	}
	if got := entries["server.host"].Source; got != "default" {
		t.Fatalf(`report source for "server.host" = %q, want %q`, got, "default")
	}
	if got := entries["database.url"].Source; got != "env:DATABASE_URL" {
		t.Fatalf(`report source for "database.url" = %q, want %q`, got, "env:DATABASE_URL")
	}
	if got := entries["database.url"].Value; got != "***" {
		t.Fatalf(`report value for "database.url" = %q, want %q`, got, "***")
	}
	if got := entries["log.level"].Source; got != path {
		t.Fatalf(`report source for "log.level" = %q, want %q`, got, path)
	}

	var out bytes.Buffer
	report.Print(&out)
	printed := out.String()
	wantParts := []string{
		"server.port",
		"source=" + path,
		"server.host",
		"source=default",
		"database.url",
		"source=env:DATABASE_URL",
		"log.level",
	}
	for _, part := range wantParts {
		if !strings.Contains(printed, part) {
			t.Fatalf("Print() output = %q, want to contain %q", printed, part)
		}
	}
}

func TestLoadReturnsReportWithError(t *testing.T) {
	type config struct {
		Name string `validate:"required"`
	}

	cfg := &config{}
	report, err := Load(cfg)
	if err == nil {
		t.Fatalf("Load() error = nil, want validation error")
	}
	if report == nil {
		t.Fatalf("Load() report = nil, want non-nil even on validation error")
	}
	if len(report.Entries) != 1 {
		t.Fatalf("len(report.Entries) = %d, want 1", len(report.Entries))
	}
	if report.Entries[0].Path != "Name" {
		t.Fatalf("report entry path = %q, want %q", report.Entries[0].Path, "Name")
	}
}
