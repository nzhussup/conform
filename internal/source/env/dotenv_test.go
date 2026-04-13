package env

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func TestDotEnvParsingHelpers(t *testing.T) {
	t.Run("parseDotEnvValue empty", func(t *testing.T) {
		got, err := parseDotEnvValue("", 1)
		if err != nil {
			t.Fatalf("parseDotEnvValue() error = %v, want nil", err)
		}
		if got != "" {
			t.Fatalf("parseDotEnvValue() = %q, want empty", got)
		}
	})

	t.Run("parseSingleQuotedDotEnvValue unterminated", func(t *testing.T) {
		_, err := parseSingleQuotedDotEnvValue("'abc", 2)
		if err == nil || !strings.Contains(err.Error(), "unterminated quoted value") {
			t.Fatalf("parseSingleQuotedDotEnvValue() error = %v, want unterminated quoted value", err)
		}
	})

	t.Run("parseSingleQuotedDotEnvValue invalid trailing", func(t *testing.T) {
		_, err := parseSingleQuotedDotEnvValue("'abc' trailing", 2)
		if err == nil || !strings.Contains(err.Error(), "invalid trailing characters") {
			t.Fatalf("parseSingleQuotedDotEnvValue() error = %v, want invalid trailing characters", err)
		}
	})

	t.Run("parseDoubleQuotedDotEnvValue invalid escape", func(t *testing.T) {
		_, err := parseDoubleQuotedDotEnvValue(`"\xZZ"`, 3)
		if err == nil || !strings.Contains(err.Error(), "invalid quoted value") {
			t.Fatalf("parseDoubleQuotedDotEnvValue() error = %v, want invalid quoted value", err)
		}
	})

	t.Run("parseDoubleQuotedDotEnvValue unterminated", func(t *testing.T) {
		_, err := parseDoubleQuotedDotEnvValue(`"unterminated`, 3)
		if err == nil || !strings.Contains(err.Error(), "unterminated quoted value") {
			t.Fatalf("parseDoubleQuotedDotEnvValue() error = %v, want unterminated quoted value", err)
		}
	})

	t.Run("parseDoubleQuotedDotEnvValue invalid trailing", func(t *testing.T) {
		_, err := parseDoubleQuotedDotEnvValue(`"abc" trailing`, 3)
		if err == nil || !strings.Contains(err.Error(), "invalid trailing characters") {
			t.Fatalf("parseDoubleQuotedDotEnvValue() error = %v, want invalid trailing characters", err)
		}
	})

	t.Run("validateDotEnvTail accepts comment and rejects garbage", func(t *testing.T) {
		if err := validateDotEnvTail("   # comment", 4); err != nil {
			t.Fatalf("validateDotEnvTail(comment) error = %v, want nil", err)
		}
		if err := validateDotEnvTail(" trailing", 4); err == nil {
			t.Fatalf("validateDotEnvTail(garbage) error = nil, want error")
		}
	})

	t.Run("parseDotEnv supports export directive", func(t *testing.T) {
		values, err := parseDotEnv([]byte("export APP_NAME=konform\n"))
		if err != nil {
			t.Fatalf("parseDotEnv() error = %v, want nil", err)
		}
		if got := values["APP_NAME"]; got != "konform" {
			t.Fatalf("APP_NAME = %q, want %q", got, "konform")
		}
	})

	t.Run("parseDotEnv rejects empty key", func(t *testing.T) {
		_, err := parseDotEnv([]byte(" =value\n"))
		if err == nil || !strings.Contains(err.Error(), "key must not be empty") {
			t.Fatalf("parseDotEnv() error = %v, want key must not be empty", err)
		}
	})

	t.Run("parseDotEnv propagates quoted value parse errors", func(t *testing.T) {
		_, err := parseDotEnv([]byte("APP='value' trailing\n"))
		if err == nil || !strings.Contains(err.Error(), "invalid trailing characters") {
			t.Fatalf("parseDotEnv() error = %v, want invalid trailing characters", err)
		}
	})

	t.Run("parseDotEnv scanner too long line", func(t *testing.T) {
		long := "KEY=" + strings.Repeat("a", 1024*1024+1) + "\n"
		_, err := parseDotEnv([]byte(long))
		if err == nil {
			t.Fatalf("parseDotEnv() error = nil, want scanner error")
		}
	})
}

func TestDotEnvFileSourceLoadFile(t *testing.T) {
	t.Run("nil schema", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		if err := os.WriteFile(path, []byte("PORT=1\n"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)

			t.Run("parseDotEnv rejects empty key", func(t *testing.T) {
				_, err := parseDotEnv([]byte(" =value\n"))
				if err == nil || !strings.Contains(err.Error(), "key must not be empty") {
					t.Fatalf("parseDotEnv() error = %v, want key must not be empty", err)
				}
			})

			t.Run("quoted value rejects invalid trailing characters", func(t *testing.T) {
				_, err := parseSingleQuotedDotEnvValue("'abc' trailing", 5)
				if err == nil {
					t.Fatalf("parseSingleQuotedDotEnvValue() error = nil, want error")
				}

				_, err = parseDoubleQuotedDotEnvValue(`"abc" trailing`, 6)
				if err == nil {
					t.Fatalf("parseDoubleQuotedDotEnvValue() error = nil, want error")
				}
			})

			t.Run("parseDotEnv scanner too long line", func(t *testing.T) {
				long := "KEY=" + strings.Repeat("a", 1024*1024+1) + "\n"
				_, err := parseDotEnv([]byte(long))
				if err == nil {
					t.Fatalf("parseDotEnv() error = nil, want scanner error")
				}
			})
		}

		source := NewDotEnvFileSource(path, "", "")
		err := source.LoadFile(nil)
		if !errors.Is(err, errs.InvalidSchemaNil) {
			t.Fatalf("LoadFile() error = %v, want wrapped %v", err, errs.InvalidSchemaNil)
		}
	})

	t.Run("read error when file is missing", func(t *testing.T) {
		source := NewDotEnvFileSource(filepath.Join(t.TempDir(), "missing.env"), "", "")
		err := source.LoadFile(&schema.Schema{})
		if err == nil {
			t.Fatalf("LoadFile() error = nil, want read error")
		}
		if !errors.Is(err, errs.DecodeSourceRead) {
			t.Fatalf("LoadFile() error = %v, want wrapped %v", err, errs.DecodeSourceRead)
		}
	})

	t.Run("loads mapped values from .env file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		content := strings.Join([]string{
			`# comment`,
			`PORT=8088`,
			`DEBUG=true`,
			`REQUEST_TIMEOUT="1500ms"`,
			`APP_NAME='konform-service'`,
			`EXTRA=ignored`,
			`INLINE=test # inline comment`,
			"",
		}, "\n")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		port := 0
		debug := false
		timeout := time.Duration(0)
		appName := ""
		inline := ""

		sc := &schema.Schema{
			Fields: []schema.Field{
				{
					Path:    "Port",
					EnvName: "PORT",
					Type:    reflect.TypeOf(0),
					Value:   reflect.ValueOf(&port).Elem(),
				},
				{
					Path:    "Debug",
					EnvName: "DEBUG",
					Type:    reflect.TypeOf(true),
					Value:   reflect.ValueOf(&debug).Elem(),
				},
				{
					Path:    "RequestTimeout",
					EnvName: "REQUEST_TIMEOUT",
					Type:    reflect.TypeOf(time.Duration(0)),
					Value:   reflect.ValueOf(&timeout).Elem(),
				},
				{
					Path:    "AppName",
					EnvName: "APP_NAME",
					Type:    reflect.TypeOf(""),
					Value:   reflect.ValueOf(&appName).Elem(),
				},
				{
					Path:    "Inline",
					EnvName: "INLINE",
					Type:    reflect.TypeOf(""),
					Value:   reflect.ValueOf(&inline).Elem(),
				},
			},
		}

		source := NewDotEnvFileSource(path, "", "")
		if err := source.LoadFile(sc); err != nil {
			t.Fatalf("LoadFile() error = %v, want nil", err)
		}

		if port != 8088 {
			t.Fatalf("Port = %d, want 8088", port)
		}
		if !debug {
			t.Fatalf("Debug = %v, want true", debug)
		}
		if timeout != 1500*time.Millisecond {
			t.Fatalf("RequestTimeout = %v, want %v", timeout, 1500*time.Millisecond)
		}
		if appName != "konform-service" {
			t.Fatalf("AppName = %q, want %q", appName, "konform-service")
		}
		if inline != "test" {
			t.Fatalf("Inline = %q, want %q", inline, "test")
		}
		if got := sc.Fields[0].Source; got != path+":PORT" {
			t.Fatalf("source = %q, want %q", got, path+":PORT")
		}
	})

	t.Run("resolves relative path using callerDir", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		if err := os.WriteFile(path, []byte("PORT=9099\n"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		port := 0
		sc := &schema.Schema{
			Fields: []schema.Field{
				{
					Path:    "Port",
					EnvName: "PORT",
					Type:    reflect.TypeOf(0),
					Value:   reflect.ValueOf(&port).Elem(),
				},
			},
		}

		source := NewDotEnvFileSource(".env", dir, "")
		if err := source.LoadFile(sc); err != nil {
			t.Fatalf("LoadFile() error = %v, want nil", err)
		}
		if port != 9099 {
			t.Fatalf("Port = %d, want 9099", port)
		}
	})

	t.Run("returns parse errors", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		if err := os.WriteFile(path, []byte("BROKEN_LINE\n"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		source := NewDotEnvFileSource(path, "", "")
		err := source.LoadFile(&schema.Schema{})
		if err == nil {
			t.Fatalf("LoadFile() error = nil, want parse error")
		}
		if !errors.Is(err, errs.DecodeSourceParse) {
			t.Fatalf("LoadFile() error = %v, want wrapped %v", err, errs.DecodeSourceParse)
		}
		if !strings.Contains(err.Error(), "expected KEY=VALUE") {
			t.Fatalf("LoadFile() error = %q, want parse details", err.Error())
		}
	})

	t.Run("returns decode errors", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		if err := os.WriteFile(path, []byte("DEBUG=not-bool\n"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		debug := false
		sc := &schema.Schema{
			Fields: []schema.Field{
				{
					Path:    "Debug",
					EnvName: "DEBUG",
					Type:    reflect.TypeOf(true),
					Value:   reflect.ValueOf(&debug).Elem(),
				},
			},
		}

		source := NewDotEnvFileSource(path, "", "")
		err := source.LoadFile(sc)
		if err == nil {
			t.Fatalf("LoadFile() error = nil, want decode error")
		}
		if !errors.Is(err, errs.Decode) {
			t.Fatalf("LoadFile() error = %v, want wrapped %v", err, errs.Decode)
		}
		if !strings.Contains(err.Error(), `env "DEBUG" -> Debug`) {
			t.Fatalf("LoadFile() error = %q, want env decode context", err.Error())
		}
	})

	t.Run("loads prefixed values from .env file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		if err := os.WriteFile(path, []byte("APP_PORT=8181\n"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		port := 0
		sc := &schema.Schema{
			Fields: []schema.Field{
				{
					Path:    "Port",
					EnvName: "PORT",
					Type:    reflect.TypeOf(0),
					Value:   reflect.ValueOf(&port).Elem(),
				},
			},
		}

		source := NewDotEnvFileSource(path, "", "APP_")
		if err := source.LoadFile(sc); err != nil {
			t.Fatalf("LoadFile() error = %v, want nil", err)
		}
		if port != 8181 {
			t.Fatalf("Port = %d, want 8181", port)
		}
		if got := sc.Fields[0].Source; got != path+":APP_PORT" {
			t.Fatalf("source = %q, want %q", got, path+":APP_PORT")
		}
	})
}
