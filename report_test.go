package konform

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	internalschema "github.com/nzhussup/konform/internal/schema"
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

func TestLoadReportPrintNoOp(t *testing.T) {
	var nilReport *LoadReport
	nilReport.Print(&bytes.Buffer{})

	empty := &LoadReport{}
	empty.Print(&bytes.Buffer{})

	withEntries := &LoadReport{Entries: []ReportEntry{{Path: "a", Value: "1", Source: "default"}}}
	withEntries.Print(nil)
}

func TestBuildReportNilSchema(t *testing.T) {
	if got := buildReport(nil); got != nil {
		t.Fatalf("buildReport(nil) = %#v, want nil", got)
	}
}

func TestBuildReportResolvesAliasesAndZeroSource(t *testing.T) {
	server := struct{}{}
	port := 8080
	dbURL := "postgres://db"

	sc := &internalschema.Schema{Fields: []internalschema.Field{
		{
			Path:    "Server",
			KeyName: "server",
			Type:    reflect.TypeOf(server),
			Value:   reflect.ValueOf(server),
		},
		{
			Path:  "Server.Port",
			Type:  reflect.TypeOf(port),
			Value: reflect.ValueOf(port),
		},
		{
			Path:    "Database.URL",
			KeyName: "database.url",
			Type:    reflect.TypeOf(dbURL),
			Value:   reflect.ValueOf(dbURL),
			Source:  "config.yaml",
		},
	}}

	report := buildReport(sc)
	if report == nil {
		t.Fatalf("buildReport() = nil, want non-nil")
	}
	if len(report.Entries) != 2 {
		t.Fatalf("len(report.Entries) = %d, want 2", len(report.Entries))
	}

	entries := make(map[string]ReportEntry, len(report.Entries))
	for _, e := range report.Entries {
		entries[e.Path] = e
	}

	portEntry, ok := entries["server.Port"]
	if !ok {
		t.Fatalf("missing alias-resolved entry for server.Port")
	}
	if portEntry.Source != "zero" {
		t.Fatalf("server.Port source = %q, want %q", portEntry.Source, "zero")
	}

	if _, ok := entries["database.url"]; !ok {
		t.Fatalf("missing entry for database.url")
	}
}

func TestResolveLookupPathForReport(t *testing.T) {
	t.Run("uses explicit key name when present", func(t *testing.T) {
		field := internalschema.Field{Path: "Server.Port", KeyName: "server_port"}
		got := resolveLookupPathForReport(field, map[string]string{"Server": "server"})
		if got != "server_port" {
			t.Fatalf("resolveLookupPathForReport() = %q, want %q", got, "server_port")
		}
	})

	t.Run("replaces full path when alias matches full field path", func(t *testing.T) {
		field := internalschema.Field{Path: "Database.URL"}
		got := resolveLookupPathForReport(field, map[string]string{"Database.URL": "database.url"})
		if got != "database.url" {
			t.Fatalf("resolveLookupPathForReport() = %q, want %q", got, "database.url")
		}
	})
}
