package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func TestLoadFile(t *testing.T) {
	makeStringField := func(path string, key string, target *string) schema.Field {
		return schema.Field{
			Path:    path,
			KeyName: key,
			Type:    reflect.TypeOf(""),
			Value:   reflect.ValueOf(target).Elem(),
		}
	}
	makeIntField := func(path string, key string, target *int) schema.Field {
		return schema.Field{
			Path:    path,
			KeyName: key,
			Type:    reflect.TypeOf(0),
			Value:   reflect.ValueOf(target).Elem(),
		}
	}

	type args struct {
		sc        *schema.Schema
		path      string
		callerDir string
		format    string
		unmarshal UnmarshalFunc
	}

	tests := []struct {
		name        string
		args        args
		setup       func(t *testing.T) args
		wantErrType error
		wantErrLike []string
		validate    func(t *testing.T, a args)
	}{
		{
			name: "nil schema",
			args: args{
				sc:        nil,
				path:      "any.yaml",
				callerDir: "",
				format:    "yaml",
				unmarshal: func(_ []byte) (Document, error) { return Document{}, nil },
			},
			wantErrType: errs.InvalidSchemaNil,
		},
		{
			name: "read file error",
			args: args{
				sc:        &schema.Schema{},
				path:      filepath.Join("does", "not", "exist.yaml"),
				callerDir: "",
				format:    "yaml",
				unmarshal: func(_ []byte) (Document, error) { return Document{}, nil },
			},
			wantErrType: errs.DecodeSourceRead,
			wantErrLike: []string{"yaml", "file"},
		},
		{
			name: "parse error from unmarshal",
			setup: func(t *testing.T) args {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "config.yaml")
				if err := os.WriteFile(p, []byte("port: ???"), 0o600); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}

				return args{
					sc:        &schema.Schema{},
					path:      p,
					callerDir: "",
					format:    "yaml",
					unmarshal: func(_ []byte) (Document, error) {
						return nil, fmt.Errorf("boom parse")
					},
				}
			},
			wantErrType: errs.DecodeSourceParse,
			wantErrLike: []string{"yaml", "boom parse"},
		},
		{
			name: "apply error from decoded field",
			setup: func(t *testing.T) args {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "config.yaml")
				if err := os.WriteFile(p, []byte("name: true"), 0o600); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}

				var name string
				sc := &schema.Schema{
					Fields: []schema.Field{
						makeStringField("Name", "name", &name),
					},
				}

				return args{
					sc:        sc,
					path:      p,
					callerDir: "",
					format:    "yaml",
					unmarshal: func(_ []byte) (Document, error) {
						return Document{"name": true}, nil
					},
				}
			},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{"yaml", "name", "Name", "expected string, got bool"},
		},
		{
			name: "relative path resolved from caller dir",
			setup: func(t *testing.T) args {
				t.Helper()
				dir := t.TempDir()
				rel := "config.yaml"
				p := filepath.Join(dir, rel)
				if err := os.WriteFile(p, []byte("port: 8080"), 0o600); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}

				var port int
				sc := &schema.Schema{
					Fields: []schema.Field{
						makeIntField("Port", "port", &port),
					},
				}

				return args{
					sc:        sc,
					path:      rel,
					callerDir: dir,
					format:    "yaml",
					unmarshal: func(_ []byte) (Document, error) {
						return Document{"port": "8080"}, nil
					},
				}
			},
			validate: func(t *testing.T, a args) {
				t.Helper()
				port := a.sc.Fields[0].Value.Interface().(int)
				if port != 8080 {
					t.Fatalf("port = %d, want 8080", port)
				}
				if got := a.sc.Fields[0].Source; got != "config.yaml" {
					t.Fatalf("field source = %q, want %q", got, "config.yaml")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.args
			if tt.setup != nil {
				a = tt.setup(t)
			}

			err := LoadFileWithMode(a.sc, a.path, a.callerDir, a.format, a.unmarshal, UnknownKeySuggestionWarn)
			if tt.wantErrType == nil {
				if err != nil {
					t.Fatalf("LoadFileWithMode() error = %v, want nil", err)
				}
				if tt.validate != nil {
					tt.validate(t, a)
				}
				return
			}

			if err == nil {
				t.Fatalf("LoadFileWithMode() error = nil, want %v", tt.wantErrType)
			}
			if !errors.Is(err, tt.wantErrType) {
				t.Fatalf("LoadFileWithMode() error = %v, want wrapped %v", err, tt.wantErrType)
			}
			for _, part := range tt.wantErrLike {
				if !strings.Contains(err.Error(), part) {
					t.Fatalf("LoadFileWithMode() error = %q, want to contain %q", err.Error(), part)
				}
			}
		})
	}
}

func TestLoadBytes(t *testing.T) {
	makeStringField := func(path string, key string, target *string) schema.Field {
		return schema.Field{
			Path:    path,
			KeyName: key,
			Type:    reflect.TypeOf(""),
			Value:   reflect.ValueOf(target).Elem(),
		}
	}
	makeIntField := func(path string, key string, target *int) schema.Field {
		return schema.Field{
			Path:    path,
			KeyName: key,
			Type:    reflect.TypeOf(0),
			Value:   reflect.ValueOf(target).Elem(),
		}
	}

	tests := []struct {
		name        string
		useWithMode bool
		scBuilder   func() *schema.Schema
		data        []byte
		format      string
		sourceLabel string
		mode        UnknownKeySuggestionMode
		unmarshal   UnmarshalFunc
		wantErrType error
		wantErrLike []string
		validate    func(t *testing.T, sc *schema.Schema)
	}{
		{
			name:        "nil schema",
			scBuilder:   func() *schema.Schema { return nil },
			data:        []byte(`{"port":"8080"}`),
			format:      "json",
			unmarshal:   func(_ []byte) (Document, error) { return Document{}, nil },
			wantErrType: errs.InvalidSchemaNil,
		},
		{
			name:      "parse error from unmarshal",
			scBuilder: func() *schema.Schema { return &schema.Schema{} },
			data:      []byte(`bad`),
			format:    "yaml",
			unmarshal: func(_ []byte) (Document, error) {
				return nil, fmt.Errorf("boom parse")
			},
			wantErrType: errs.DecodeSourceParse,
			wantErrLike: []string{"yaml", "boom parse"},
		},
		{
			name: "apply error from decoded field",
			scBuilder: func() *schema.Schema {
				var name string
				return &schema.Schema{
					Fields: []schema.Field{
						makeStringField("Name", "name", &name),
					},
				}
			},
			data:   []byte(`{"name":true}`),
			format: "json",
			unmarshal: func(_ []byte) (Document, error) {
				return Document{"name": true}, nil
			},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{"json", "name", "Name", "expected string, got bool"},
		},
		{
			name: "success with default source label",
			scBuilder: func() *schema.Schema {
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						makeIntField("Port", "port", &port),
					},
				}
			},
			data:   []byte(`{"port":"8080"}`),
			format: "json",
			unmarshal: func(_ []byte) (Document, error) {
				return Document{"port": "8080"}, nil
			},
			validate: func(t *testing.T, sc *schema.Schema) {
				t.Helper()
				if got := sc.Fields[0].Value.Interface().(int); got != 8080 {
					t.Fatalf("port = %d, want 8080", got)
				}
				if got := sc.Fields[0].Source; got != "json" {
					t.Fatalf("field source = %q, want %q", got, "json")
				}
			},
		},
		{
			name:        "success with explicit source label and strict unknown-key mode",
			useWithMode: true,
			scBuilder: func() *schema.Schema {
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						makeIntField("Port", "port", &port),
					},
				}
			},
			data:        []byte(`{"port":"8080"}`),
			format:      "json",
			mode:        UnknownKeySuggestionError,
			sourceLabel: "inline-json",
			unmarshal: func(_ []byte) (Document, error) {
				return Document{"port": "8080"}, nil
			},
			validate: func(t *testing.T, sc *schema.Schema) {
				t.Helper()
				if got := sc.Fields[0].Value.Interface().(int); got != 8080 {
					t.Fatalf("port = %d, want 8080", got)
				}
				if got := sc.Fields[0].Source; got != "inline-json" {
					t.Fatalf("field source = %q, want %q", got, "inline-json")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := tt.scBuilder()

			mode := UnknownKeySuggestionWarn
			sourceLabel := tt.format
			if tt.useWithMode {
				mode = tt.mode
				sourceLabel = tt.sourceLabel
			}
			err := LoadBytesWithMode(sc, tt.data, tt.format, tt.unmarshal, mode, sourceLabel)

			if tt.wantErrType == nil {
				if err != nil {
					t.Fatalf("LoadBytesWithMode() error = %v, want nil", err)
				}
				if tt.validate != nil {
					tt.validate(t, sc)
				}
				return
			}

			if err == nil {
				t.Fatalf("LoadBytesWithMode() error = nil, want %v", tt.wantErrType)
			}
			if !errors.Is(err, tt.wantErrType) {
				t.Fatalf("LoadBytesWithMode() error = %v, want wrapped %v", err, tt.wantErrType)
			}
			for _, part := range tt.wantErrLike {
				if !strings.Contains(err.Error(), part) {
					t.Fatalf("LoadBytesWithMode() error = %q, want to contain %q", err.Error(), part)
				}
			}
		})
	}
}
