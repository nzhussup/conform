package konform

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	_, err := Load(target)
	if !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("Load() error = %v, want wrapped %v", err, ErrInvalidTarget)
	}
}

func TestLoadOptionError(t *testing.T) {
	cfg := &loadTestConfig{}
	_, err := Load(cfg, func(_ *loadOptions) error { return fmt.Errorf("option failed") })
	if err == nil || !strings.Contains(err.Error(), "option failed") {
		t.Fatalf("Load() error = %v, want to contain %q", err, "option failed")
	}
}

func TestLoadSourceError(t *testing.T) {
	cfg := &loadTestConfig{}
	_, err := Load(cfg, optionWithSource(func(_ *internalschema.Schema) error {
		return fmt.Errorf("source failed")
	}))
	if err == nil || !strings.Contains(err.Error(), "source failed") {
		t.Fatalf("Load() error = %v, want to contain %q", err, "source failed")
	}
}

func TestLoadValidationErrorForMissingRequired(t *testing.T) {
	cfg := &loadTestConfig{}
	_, err := Load(cfg)

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
	_, err := Load(cfg, optionWithSource(func(sc *internalschema.Schema) error {
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
	_, err := Load(cfg, FromEnv())
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

func TestLoadSuccessWithDotEnvFileSource(t *testing.T) {
	cfg := &loadTestConfig{}

	dir := t.TempDir()
	dotEnvPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(dotEnvPath, []byte("NAME=api-dotenv\nPORT=7070\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(cfg, FromDotEnvFile(dotEnvPath))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Name != "api-dotenv" {
		t.Fatalf("Name = %q, want %q", cfg.Name, "api-dotenv")
	}
	if cfg.Port != 7070 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 7070)
	}
}

func TestLoadSuccessWithEnvPrefix(t *testing.T) {
	t.Setenv("APP_NAME", "api-prefixed")
	t.Setenv("APP_PORT", "7071")

	cfg := &loadTestConfig{}
	_, err := Load(cfg, FromEnv(), WithEnvPrefix("APP_"))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Name != "api-prefixed" {
		t.Fatalf("Name = %q, want %q", cfg.Name, "api-prefixed")
	}
	if cfg.Port != 7071 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 7071)
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
	_, err := Load(cfg, FromJSONFile(path))
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

func TestLoadUnknownKeySuggestionMode(t *testing.T) {
	type config struct {
		AppName string `validate:"required"`
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"App":{"Name":"konform-service"}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Run("default mode warns and falls through to validation", func(t *testing.T) {
		cfg := &config{}
		stderr := captureStderr(t, func() {
			_, err := Load(cfg, FromJSONFile(path))
			if err == nil {
				t.Fatalf("Load() error = nil, want validation error")
			}
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("Load() error = %v, want wrapped %v", err, ErrValidation)
			}
			if !strings.Contains(err.Error(), "AppName: required") {
				t.Fatalf("Load() error = %q, want required validation for AppName", err.Error())
			}
		})
		if !strings.Contains(stderr, `warning: json: unknown configuration key "App.Name"`) {
			t.Fatalf("stderr = %q, want warning message", stderr)
		}
	})

	t.Run("error mode reports unexpected input key with schema suggestion", func(t *testing.T) {
		cfg := &config{}
		_, err := Load(cfg, FromJSONFile(path), WithUnknownKeySuggestionMode(Error))
		if err == nil {
			t.Fatalf("Load() error = nil, want decode error")
		}
		if !errors.Is(err, ErrDecode) {
			t.Fatalf("Load() error = %v, want wrapped %v", err, ErrDecode)
		}
		wantParts := []string{`unknown configuration key "App.Name"`, `did you mean "AppName"?`}
		for _, part := range wantParts {
			if !strings.Contains(err.Error(), part) {
				t.Fatalf("Load() error = %q, want to contain %q", err.Error(), part)
			}
		}
	})

	t.Run("off mode ignores unknown keys and falls through to validation", func(t *testing.T) {
		cfg := &config{}
		stderr := captureStderr(t, func() {
			_, err := Load(cfg, FromJSONFile(path), WithUnknownKeySuggestionMode(Off))
			if err == nil {
				t.Fatalf("Load() error = nil, want validation error")
			}
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("Load() error = %v, want wrapped %v", err, ErrValidation)
			}
		})
		if strings.Contains(stderr, "unknown configuration key") {
			t.Fatalf("stderr = %q, want no unknown-key warnings", stderr)
		}
	})

	t.Run("off mode works even when option is set before source", func(t *testing.T) {
		cfg := &config{}
		_, err := Load(cfg, WithUnknownKeySuggestionMode(Off), FromJSONFile(path))
		if err == nil {
			t.Fatalf("Load() error = nil, want validation error")
		}
		if !errors.Is(err, ErrValidation) {
			t.Fatalf("Load() error = %v, want wrapped %v", err, ErrValidation)
		}
	})

	t.Run("strict mode reports unknown file key as decode error", func(t *testing.T) {
		cfg := &config{}
		_, err := Load(cfg, Strict(), FromJSONFile(path))
		if err == nil {
			t.Fatalf("Load() error = nil, want decode error")
		}
		if !errors.Is(err, ErrDecode) {
			t.Fatalf("Load() error = %v, want wrapped %v", err, ErrDecode)
		}
		if !strings.Contains(err.Error(), `unknown configuration key "App.Name"`) {
			t.Fatalf("Load() error = %q, want unknown file key error", err.Error())
		}
	})
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	originalStderr := os.Stderr
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = writePipe

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, readPipe)
		done <- buf.String()
	}()

	fn()

	_ = writePipe.Close()
	os.Stderr = originalStderr
	output := <-done
	_ = readPipe.Close()
	return output
}

func TestLoadStrictMode(t *testing.T) {
	t.Run("missing optional field does not fail", func(t *testing.T) {
		type config struct {
			License string
		}

		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		if err := os.WriteFile(path, []byte(`{}`), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		cfg := &config{}
		if _, err := Load(cfg, Strict(), FromJSONFile(path)); err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}
		if cfg.License != "" {
			t.Fatalf("License = %q, want empty", cfg.License)
		}
	})

	t.Run("missing required field fails", func(t *testing.T) {
		type config struct {
			License string `validate:"required"`
		}

		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		if err := os.WriteFile(path, []byte(`{}`), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		cfg := &config{}
		_, err := Load(cfg, Strict(), FromJSONFile(path))
		if err == nil {
			t.Fatalf("Load() error = nil, want validation error")
		}
		if !errors.Is(err, ErrValidation) {
			t.Fatalf("Load() error = %v, want wrapped %v", err, ErrValidation)
		}
	})

	t.Run("strict mode rejects unknown structured input keys", func(t *testing.T) {
		type config struct {
			Database struct {
				Host string
			}
		}

		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		if err := os.WriteFile(path, []byte(`{"Databas":{"Host":"localhost"}}`), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		cfg := &config{}
		_, err := Load(cfg, Strict(), FromJSONFile(path))
		if err == nil {
			t.Fatalf("Load() error = nil, want decode error")
		}
		if !errors.Is(err, ErrDecode) {
			t.Fatalf("Load() error = %v, want wrapped %v", err, ErrDecode)
		}
		if !strings.Contains(err.Error(), `unknown configuration key "Databas.Host"`) {
			t.Fatalf("Load() error = %q, want unexpected key details", err.Error())
		}
		if !strings.Contains(err.Error(), `did you mean "Database.Host"?`) {
			t.Fatalf("Load() error = %q, want schema suggestion", err.Error())
		}
	})

	t.Run("strict mode rejects duplicate key mappings", func(t *testing.T) {
		type config struct {
			A string `key:"app.name"`
			B string `key:"app.name"`
		}

		cfg := &config{}
		_, err := Load(cfg, Strict())
		if err == nil {
			t.Fatalf("Load() error = nil, want invalid schema error")
		}
		if !errors.Is(err, ErrInvalidSchema) {
			t.Fatalf("Load() error = %v, want wrapped %v", err, ErrInvalidSchema)
		}
		if !strings.Contains(err.Error(), `conflicting key mapping "app.name"`) {
			t.Fatalf("Load() error = %q, want conflicting key mapping details", err.Error())
		}
	})

	t.Run("strict mode rejects duplicate env mappings", func(t *testing.T) {
		type config struct {
			A string `env:"APP_NAME"`
			B string `env:"APP_NAME"`
		}

		cfg := &config{}
		_, err := Load(cfg, Strict())
		if err == nil {
			t.Fatalf("Load() error = nil, want invalid schema error")
		}
		if !errors.Is(err, ErrInvalidSchema) {
			t.Fatalf("Load() error = %v, want wrapped %v", err, ErrInvalidSchema)
		}
		if !strings.Contains(err.Error(), `conflicting env mapping "APP_NAME"`) {
			t.Fatalf("Load() error = %q, want conflicting env mapping details", err.Error())
		}
	})

	t.Run("strict mode keeps env decode errors", func(t *testing.T) {
		type config struct {
			Port int `env:"PORT"`
		}

		t.Setenv("PORT", "bad-int")
		cfg := &config{}
		_, err := Load(cfg, Strict(), FromEnv())
		if err == nil {
			t.Fatalf("Load() error = nil, want decode error")
		}
		if !errors.Is(err, ErrDecode) {
			t.Fatalf("Load() error = %v, want wrapped %v", err, ErrDecode)
		}
		if !strings.Contains(err.Error(), "invalid int value") {
			t.Fatalf("Load() error = %q, want invalid int decode details", err.Error())
		}
	})

	t.Run("strict mode ignores unrelated env vars", func(t *testing.T) {
		type config struct {
			Port int `env:"PORT"`
		}

		t.Setenv("PORT", "8080")
		t.Setenv("UNRELATED_KEY", "ignored")

		cfg := &config{}
		_, err := Load(cfg, Strict(), FromEnv())
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}
		if cfg.Port != 8080 {
			t.Fatalf("Port = %d, want 8080", cfg.Port)
		}
	})

	t.Run("strict mode keeps file decode mismatch errors", func(t *testing.T) {
		type config struct {
			Port int `key:"port"`
		}

		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		if err := os.WriteFile(path, []byte(`{"port":"bad-int"}`), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		cfg := &config{}
		_, err := Load(cfg, Strict(), FromJSONFile(path))
		if err == nil {
			t.Fatalf("Load() error = nil, want decode error")
		}
		if !errors.Is(err, ErrDecode) {
			t.Fatalf("Load() error = %v, want wrapped %v", err, ErrDecode)
		}
		if !strings.Contains(err.Error(), "invalid int value") {
			t.Fatalf("Load() error = %q, want invalid int decode details", err.Error())
		}
	})
}
