package konform

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (*loadOptions, *schema.Schema)
		wantErrType error
		validate    func(t *testing.T, o *loadOptions, sc *schema.Schema)
	}{
		{
			name: "nil load options",
			setup: func(t *testing.T) (*loadOptions, *schema.Schema) {
				t.Helper()
				return nil, nil
			},
			wantErrType: errs.InvalidSchemaNilOptions,
		},
		{
			name: "registers env source and loads value",
			setup: func(t *testing.T) (*loadOptions, *schema.Schema) {
				t.Helper()
				t.Setenv("PORT", "9090")

				var port int
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
				return &loadOptions{}, sc
			},
			validate: func(t *testing.T, o *loadOptions, sc *schema.Schema) {
				t.Helper()
				if got := len(o.sources); got != 1 {
					t.Fatalf("len(sources) = %d, want 1", got)
				}
				if err := o.sources[0](sc); err != nil {
					t.Fatalf("source() error = %v, want nil", err)
				}
				if got := sc.Fields[0].Value.Interface().(int); got != 9090 {
					t.Fatalf("Port = %d, want 9090", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, sc := tt.setup(t)
			err := FromEnv()(o)
			if tt.wantErrType != nil {
				if err == nil {
					t.Fatalf("FromEnv() error = nil, want %v", tt.wantErrType)
				}
				if !errors.Is(err, tt.wantErrType) {
					t.Fatalf("FromEnv() error = %v, want wrapped %v", err, tt.wantErrType)
				}
				return
			}
			if err != nil {
				t.Fatalf("FromEnv() error = %v, want nil", err)
			}
			if tt.validate != nil {
				tt.validate(t, o, sc)
			}
		})
	}
}

func TestFileOptions(t *testing.T) {
	makeSchema := func(target *int) *schema.Schema {
		return &schema.Schema{
			Fields: []schema.Field{
				{
					Path:    "Port",
					KeyName: "port",
					Type:    reflect.TypeOf(0),
					Value:   reflect.ValueOf(target).Elem(),
				},
			},
		}
	}

	tests := []struct {
		name        string
		option      func(path string) Option
		ext         string
		path        string
		content     string
		wantErrType error
		validate    func(t *testing.T, o *loadOptions, sc *schema.Schema)
	}{
		{
			name:        "yaml empty path",
			option:      FromYAMLFile,
			ext:         ".yaml",
			path:        "",
			wantErrType: errs.InvalidSchemaEmptyYAML,
		},
		{
			name:        "json empty path",
			option:      FromJSONFile,
			ext:         ".json",
			path:        "",
			wantErrType: errs.InvalidSchemaEmptyJSON,
		},
		{
			name:    "yaml registers source and loads absolute path",
			option:  FromYAMLFile,
			ext:     ".yaml",
			content: "port: \"8081\"\n",
			validate: func(t *testing.T, o *loadOptions, sc *schema.Schema) {
				t.Helper()
				if got := len(o.sources); got != 1 {
					t.Fatalf("len(sources) = %d, want 1", got)
				}
				if err := o.sources[0](sc); err != nil {
					t.Fatalf("source() error = %v, want nil", err)
				}
				if got := sc.Fields[0].Value.Interface().(int); got != 8081 {
					t.Fatalf("Port = %d, want 8081", got)
				}
			},
		},
		{
			name:    "json registers source and loads absolute path",
			option:  FromJSONFile,
			ext:     ".json",
			content: `{"port":"8082"}`,
			validate: func(t *testing.T, o *loadOptions, sc *schema.Schema) {
				t.Helper()
				if got := len(o.sources); got != 1 {
					t.Fatalf("len(sources) = %d, want 1", got)
				}
				if err := o.sources[0](sc); err != nil {
					t.Fatalf("source() error = %v, want nil", err)
				}
				if got := sc.Fields[0].Value.Interface().(int); got != 8082 {
					t.Fatalf("Port = %d, want 8082", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path
			if tt.content != "" {
				dir := t.TempDir()
				path = filepath.Join(dir, "config"+tt.ext)
				if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			}

			opt := tt.option(path)
			err := opt(nil)
			if tt.wantErrType != nil {
				if err == nil {
					t.Fatalf("option(nil) error = nil, want %v", tt.wantErrType)
				}
				if !errors.Is(err, tt.wantErrType) {
					t.Fatalf("option(nil) error = %v, want wrapped %v", err, tt.wantErrType)
				}
				return
			}
			if !errors.Is(err, errs.InvalidSchemaNilOptions) {
				t.Fatalf("option(nil) error = %v, want wrapped %v", err, errs.InvalidSchemaNilOptions)
			}

			o := &loadOptions{}
			if err := opt(o); err != nil {
				t.Fatalf("option(loadOptions) error = %v, want nil", err)
			}

			var port int
			sc := makeSchema(&port)
			if tt.validate != nil {
				tt.validate(t, o, sc)
			}
		})
	}
}
