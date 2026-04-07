package yaml

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func TestNewFileSource(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		callerDir string
	}{
		{name: "regular values", path: "config.yaml", callerDir: "/tmp"},
		{name: "empty values", path: "", callerDir: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFileSource(tt.path, tt.callerDir)
			if got.path != tt.path {
				t.Fatalf("path = %q, want %q", got.path, tt.path)
			}
			if got.callerDir != tt.callerDir {
				t.Fatalf("callerDir = %q, want %q", got.callerDir, tt.callerDir)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	makeSchema := func(port *int, mode *string) *schema.Schema {
		fields := []schema.Field{
			{
				Path:    "Port",
				KeyName: "port",
				Type:    reflect.TypeOf(0),
				Value:   reflect.ValueOf(port).Elem(),
			},
		}
		if mode != nil {
			fields = append(fields, schema.Field{
				Path:    "Mode",
				KeyName: "mode",
				Type:    reflect.TypeOf(""),
				Value:   reflect.ValueOf(mode).Elem(),
			})
		}
		return &schema.Schema{Fields: fields}
	}

	tests := []struct {
		name        string
		setup       func(t *testing.T) (FileSource, *schema.Schema)
		wantErrType error
		wantErrLike []string
		validate    func(t *testing.T, sc *schema.Schema)
	}{
		{
			name: "nil schema",
			setup: func(t *testing.T) (FileSource, *schema.Schema) {
				t.Helper()
				return NewFileSource("config.yaml", ""), nil
			},
			wantErrType: errs.InvalidSchemaNil,
		},
		{
			name: "missing file",
			setup: func(t *testing.T) (FileSource, *schema.Schema) {
				t.Helper()
				var port int
				return NewFileSource("missing.yaml", t.TempDir()), makeSchema(&port, nil)
			},
			wantErrType: errs.DecodeSourceRead,
		},
		{
			name: "parse error",
			setup: func(t *testing.T) (FileSource, *schema.Schema) {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "config.yaml")
				if err := os.WriteFile(p, []byte("port: ["), 0o600); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
				var port int
				return NewFileSource("config.yaml", dir), makeSchema(&port, nil)
			},
			wantErrType: errs.DecodeSourceParse,
		},
		{
			name: "decode field error",
			setup: func(t *testing.T) (FileSource, *schema.Schema) {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "config.yaml")
				content := "port: \"8080\"\nmode: true\n"
				if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
				var port int
				var mode string
				return NewFileSource("config.yaml", dir), makeSchema(&port, &mode)
			},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{`yaml "mode" -> Mode`, "expected string, got bool"},
		},
		{
			name: "success",
			setup: func(t *testing.T) (FileSource, *schema.Schema) {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "config.yaml")
				if err := os.WriteFile(p, []byte("port: \"8080\"\n"), 0o600); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
				var port int
				return NewFileSource("config.yaml", dir), makeSchema(&port, nil)
			},
			validate: func(t *testing.T, sc *schema.Schema) {
				t.Helper()
				if got := sc.Fields[0].Value.Interface().(int); got != 8080 {
					t.Fatalf("Port = %d, want 8080", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, sc := tt.setup(t)
			err := source.Load(sc)
			if tt.wantErrType != nil {
				if err == nil {
					t.Fatalf("Load() error = nil, want %v", tt.wantErrType)
				}
				if !errors.Is(err, tt.wantErrType) {
					t.Fatalf("Load() error = %v, want wrapped %v", err, tt.wantErrType)
				}
				for _, part := range tt.wantErrLike {
					if !strings.Contains(err.Error(), part) {
						t.Fatalf("Load() error = %q, want to contain %q", err.Error(), part)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
			if tt.validate != nil {
				tt.validate(t, sc)
			}
		})
	}
}
