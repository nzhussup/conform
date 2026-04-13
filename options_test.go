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
		{
			name: "loads prefixed env values when WithEnvPrefix is set",
			setup: func(t *testing.T) (*loadOptions, *schema.Schema) {
				t.Helper()
				t.Setenv("APP_PORT", "9091")

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
				o := &loadOptions{}
				if err := WithEnvPrefix("APP_")(o); err != nil {
					t.Fatalf("WithEnvPrefix() error = %v, want nil", err)
				}
				return o, sc
			},
			validate: func(t *testing.T, o *loadOptions, sc *schema.Schema) {
				t.Helper()
				if err := o.sources[0](sc); err != nil {
					t.Fatalf("source() error = %v, want nil", err)
				}
				if got := sc.Fields[0].Value.Interface().(int); got != 9091 {
					t.Fatalf("Port = %d, want 9091", got)
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

func TestFromDotEnvFile(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		assertOptionError(t, FromDotEnvFile("")(nil), errs.InvalidSchemaEmptyDotEnv)
	})

	t.Run("nil load options", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		if err := os.WriteFile(path, []byte("PORT=8084\n"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		assertOptionError(t, FromDotEnvFile(path)(nil), errs.InvalidSchemaNilOptions)
	})

	t.Run("registers .env file source and loads value", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		if err := os.WriteFile(path, []byte("PORT=8084\n"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		o := &loadOptions{}
		if err := FromDotEnvFile(path)(o); err != nil {
			t.Fatalf("FromDotEnvFile() error = %v, want nil", err)
		}
		if got := len(o.sources); got != 1 {
			t.Fatalf("len(sources) = %d, want 1", got)
		}

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
		if err := o.sources[0](sc); err != nil {
			t.Fatalf("source() error = %v, want nil", err)
		}
		if port != 8084 {
			t.Fatalf("Port = %d, want 8084", port)
		}
	})

	t.Run("loads prefixed .env value when WithEnvPrefix is set", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		if err := os.WriteFile(path, []byte("APP_PORT=8085\n"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		o := &loadOptions{}
		if err := FromDotEnvFile(path)(o); err != nil {
			t.Fatalf("FromDotEnvFile() error = %v, want nil", err)
		}
		if err := WithEnvPrefix("APP_")(o); err != nil {
			t.Fatalf("WithEnvPrefix() error = %v, want nil", err)
		}

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

		if err := o.sources[0](sc); err != nil {
			t.Fatalf("source() error = %v, want nil", err)
		}
		if port != 8085 {
			t.Fatalf("Port = %d, want 8085", port)
		}
	})
}

func TestWithEnvPrefix(t *testing.T) {
	t.Run("nil load options", func(t *testing.T) {
		err := WithEnvPrefix("APP_")(nil)
		if !errors.Is(err, errs.InvalidSchemaNilOptions) {
			t.Fatalf("WithEnvPrefix(nil) error = %v, want %v", err, errs.InvalidSchemaNilOptions)
		}
	})

	t.Run("sets prefix", func(t *testing.T) {
		o := &loadOptions{}
		if err := WithEnvPrefix("APP_")(o); err != nil {
			t.Fatalf("WithEnvPrefix() error = %v, want nil", err)
		}
		if got := o.envPrefix; got != "APP_" {
			t.Fatalf("envPrefix = %q, want %q", got, "APP_")
		}
	})
}

func TestWithCustomValidator(t *testing.T) {
	t.Run("nil load options", func(t *testing.T) {
		err := WithCustomValidator("custom", func(_ any, _ string) error { return nil })(nil)
		if !errors.Is(err, errs.InvalidSchemaNilOptions) {
			t.Fatalf("WithCustomValidator(nil) error = %v, want %v", err, errs.InvalidSchemaNilOptions)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		err := WithCustomValidator("", func(_ any, _ string) error { return nil })(&loadOptions{})
		if !errors.Is(err, errs.InvalidSchema) {
			t.Fatalf("WithCustomValidator(empty) error = %v, want wrapped %v", err, errs.InvalidSchema)
		}
	})

	t.Run("nil function", func(t *testing.T) {
		err := WithCustomValidator("custom", nil)(&loadOptions{})
		if !errors.Is(err, errs.InvalidSchema) {
			t.Fatalf("WithCustomValidator(nil fn) error = %v, want wrapped %v", err, errs.InvalidSchema)
		}
	})

	t.Run("registers validator", func(t *testing.T) {
		o := &loadOptions{}
		validator := func(_ any, _ string) error { return nil }
		if err := WithCustomValidator("custom", validator)(o); err != nil {
			t.Fatalf("WithCustomValidator() error = %v, want nil", err)
		}
		if o.customValidators == nil {
			t.Fatalf("customValidators = nil, want map")
		}
		got, ok := o.customValidators["custom"]
		if !ok {
			t.Fatalf("customValidators missing %q", "custom")
		}
		if reflect.ValueOf(got).Pointer() != reflect.ValueOf(validator).Pointer() {
			t.Fatalf("registered validator function does not match input function")
		}
	})
}

func TestUnknownKeySuggestionOptions(t *testing.T) {
	t.Run("nil load options", func(t *testing.T) {
		err := WithUnknownKeySuggestionMode(ModeOff)(nil)
		if !errors.Is(err, errs.InvalidSchemaNilOptions) {
			t.Fatalf("WithUnknownKeySuggestionMode(nil) error = %v, want %v", err, errs.InvalidSchemaNilOptions)
		}
	})

	t.Run("sets explicit error mode", func(t *testing.T) {
		o := &loadOptions{}
		if err := WithUnknownKeySuggestionMode(ModeError)(o); err != nil {
			t.Fatalf("WithUnknownKeySuggestionMode() error = %v, want nil", err)
		}
		if o.unknownKeySuggestMode != ModeError {
			t.Fatalf("unknownKeySuggestMode = %v, want %v", o.unknownKeySuggestMode, ModeError)
		}
	})

	t.Run("sets explicit off mode", func(t *testing.T) {
		o := &loadOptions{}
		if err := WithUnknownKeySuggestionMode(ModeOff)(o); err != nil {
			t.Fatalf("WithUnknownKeySuggestionMode() error = %v, want nil", err)
		}
		if o.unknownKeySuggestMode != ModeOff {
			t.Fatalf("unknownKeySuggestMode = %v, want %v", o.unknownKeySuggestMode, ModeOff)
		}
	})

	t.Run("default zero-value mode is warn", func(t *testing.T) {
		o := &loadOptions{}
		if o.unknownKeySuggestMode != ModeWarn {
			t.Fatalf("unknownKeySuggestMode = %v, want %v", o.unknownKeySuggestMode, ModeWarn)
		}
	})
}

func TestStrictOption(t *testing.T) {
	t.Run("nil load options", func(t *testing.T) {
		err := Strict()(nil)
		if !errors.Is(err, errs.InvalidSchemaNilOptions) {
			t.Fatalf("Strict()(nil) error = %v, want %v", err, errs.InvalidSchemaNilOptions)
		}
	})

	t.Run("sets strict mode", func(t *testing.T) {
		o := &loadOptions{}
		if err := Strict()(o); err != nil {
			t.Fatalf("Strict() error = %v, want nil", err)
		}
		if !o.strict {
			t.Fatalf("strict = false, want true")
		}
	})
}

func TestFileOptions(t *testing.T) {
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
			name:        "toml empty path",
			option:      FromTOMLFile,
			ext:         ".toml",
			path:        "",
			wantErrType: errs.InvalidSchemaEmptyTOML,
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
		{
			name:    "toml registers source and loads absolute path",
			option:  FromTOMLFile,
			ext:     ".toml",
			content: "port = \"8083\"\n",
			validate: func(t *testing.T, o *loadOptions, sc *schema.Schema) {
				t.Helper()
				if got := len(o.sources); got != 1 {
					t.Fatalf("len(sources) = %d, want 1", got)
				}
				if err := o.sources[0](sc); err != nil {
					t.Fatalf("source() error = %v, want nil", err)
				}
				if got := sc.Fields[0].Value.Interface().(int); got != 8083 {
					t.Fatalf("Port = %d, want 8083", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createOptionPath(t, tt.path, tt.ext, tt.content)

			opt := tt.option(path)
			if tt.wantErrType != nil {
				assertOptionError(t, opt(nil), tt.wantErrType)
				return
			}
			assertOptionError(t, opt(nil), errs.InvalidSchemaNilOptions)

			o := &loadOptions{}
			if err := opt(o); err != nil {
				t.Fatalf("option(loadOptions) error = %v, want nil", err)
			}

			var port int
			sc := makePortSchema(&port)
			if tt.validate != nil {
				tt.validate(t, o, sc)
			}
		})
	}
}

func TestBytesOptions(t *testing.T) {
	tests := []struct {
		name        string
		option      func(data []byte) Option
		data        []byte
		wantErrType error
		validate    func(t *testing.T, o *loadOptions, sc *schema.Schema)
	}{
		{
			name:        "yaml empty bytes",
			option:      FromYAMLBytes,
			data:        nil,
			wantErrType: errs.InvalidSchemaEmptyYAMLBytes,
		},
		{
			name:        "json empty bytes",
			option:      FromJSONBytes,
			data:        nil,
			wantErrType: errs.InvalidSchemaEmptyJSONBytes,
		},
		{
			name:        "toml empty bytes",
			option:      FromTOMLBytes,
			data:        nil,
			wantErrType: errs.InvalidSchemaEmptyTOMLBytes,
		},
		{
			name:   "yaml registers source and loads data",
			option: FromYAMLBytes,
			data:   []byte("port: \"8081\"\n"),
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
			name:   "json registers source and loads data",
			option: FromJSONBytes,
			data:   []byte(`{"port":"8082"}`),
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
		{
			name:   "toml registers source and loads data",
			option: FromTOMLBytes,
			data:   []byte("port = \"8083\"\n"),
			validate: func(t *testing.T, o *loadOptions, sc *schema.Schema) {
				t.Helper()
				if got := len(o.sources); got != 1 {
					t.Fatalf("len(sources) = %d, want 1", got)
				}
				if err := o.sources[0](sc); err != nil {
					t.Fatalf("source() error = %v, want nil", err)
				}
				if got := sc.Fields[0].Value.Interface().(int); got != 8083 {
					t.Fatalf("Port = %d, want 8083", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := tt.option(tt.data)
			if tt.wantErrType != nil {
				assertOptionError(t, opt(nil), tt.wantErrType)
				return
			}
			assertOptionError(t, opt(nil), errs.InvalidSchemaNilOptions)

			o := &loadOptions{}
			if err := opt(o); err != nil {
				t.Fatalf("option(loadOptions) error = %v, want nil", err)
			}

			var port int
			sc := makePortSchema(&port)
			if tt.validate != nil {
				tt.validate(t, o, sc)
			}
		})
	}
}

func makePortSchema(target *int) *schema.Schema {
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

func createOptionPath(t *testing.T, explicitPath string, ext string, content string) string {
	t.Helper()
	if content == "" {
		return explicitPath
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config"+ext)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func assertOptionError(t *testing.T, err error, wantErr error) {
	t.Helper()
	if err == nil {
		t.Fatalf("option(nil) error = nil, want %v", wantErr)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("option(nil) error = %v, want wrapped %v", err, wantErr)
	}
}
