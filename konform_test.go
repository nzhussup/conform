package konform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/decode"
	internalschema "github.com/nzhussup/konform/internal/schema"
)

type loadTestConfig struct {
	Name string `env:"NAME" validate:"required"`
	Port int    `default:"8080" env:"PORT"`
}

func optionWithSource(loader sourceLoader) Option {
	return func(o *loadOptions) error {
		if o == nil {
			return ErrInvalidSchema
		}
		o.sources = append(o.sources, loader)
		return nil
	}
}

func TestLoadInvalidTarget(t *testing.T) {
	var target loadTestConfig
	err := Load(target)
	if !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("Load() error = %v, want wrapped %v", err, ErrInvalidTarget)
	}
}

func TestLoadOptionError(t *testing.T) {
	cfg := &loadTestConfig{}
	err := Load(cfg, func(_ *loadOptions) error { return fmt.Errorf("option failed") })
	if err == nil || !strings.Contains(err.Error(), "option failed") {
		t.Fatalf("Load() error = %v, want to contain %q", err, "option failed")
	}
}

func TestLoadSourceError(t *testing.T) {
	cfg := &loadTestConfig{}
	err := Load(cfg, optionWithSource(func(_ *internalschema.Schema) error {
		return fmt.Errorf("source failed")
	}))
	if err == nil || !strings.Contains(err.Error(), "source failed") {
		t.Fatalf("Load() error = %v, want to contain %q", err, "source failed")
	}
}

func TestLoadValidationErrorForMissingRequired(t *testing.T) {
	cfg := &loadTestConfig{}
	err := Load(cfg)

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("Load() error = %v, want wrapped %v", err, ErrValidation)
	}

	var vErr *ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("Load() error = %v, want ValidationError", err)
	}
	if !strings.Contains(err.Error(), "Name: required") {
		t.Fatalf("Load() error = %q, want to contain %q", err.Error(), "Name: required")
	}
}

func TestLoadSuccessWithDefaultThenSourceOverride(t *testing.T) {
	cfg := &loadTestConfig{}
	err := Load(cfg, optionWithSource(func(sc *internalschema.Schema) error {
		for _, f := range sc.Fields {
			if f.Path == "Name" {
				if err := decode.SetFieldValue(f, "svc"); err != nil {
					return err
				}
				continue
			}
			if f.Path == "Port" {
				if err := decode.SetFieldValue(f, "9090"); err != nil {
					return err
				}
			}
		}
		return nil
	}))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Name != "svc" {
		t.Fatalf("Name = %q, want %q", cfg.Name, "svc")
	}
	if cfg.Port != 9090 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 9090)
	}
}

func TestLoadSuccessWithEnvSource(t *testing.T) {
	t.Setenv("NAME", "api")
	t.Setenv("PORT", "7777")

	cfg := &loadTestConfig{}
	err := Load(cfg, FromEnv())
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Name != "api" {
		t.Fatalf("Name = %q, want %q", cfg.Name, "api")
	}
	if cfg.Port != 7777 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 7777)
	}
}

func TestLoadReportsMultipleDecodeErrorsFromFile(t *testing.T) {
	type config struct {
		Name  string `key:"name"`
		Port  int    `key:"port"`
		Debug bool   `key:"debug"`
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"name":true,"port":"not-int","debug":"not-bool"}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &config{}
	err := Load(cfg, FromJSONFile(path))
	if err == nil {
		t.Fatalf("Load() error = nil, want decode errors")
	}
	if !errors.Is(err, ErrDecode) {
		t.Fatalf("Load() error = %v, want wrapped %v", err, ErrDecode)
	}

	wantParts := []string{
		`json "name" -> Name`,
		"expected string, got bool",
		`json "port" -> Port`,
		"invalid int value",
		`json "debug" -> Debug`,
		"invalid bool value",
	}
	for _, part := range wantParts {
		if !strings.Contains(err.Error(), part) {
			t.Fatalf("Load() error = %q, want to contain %q", err.Error(), part)
		}
	}
}
