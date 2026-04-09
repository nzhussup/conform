package json

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
	"github.com/nzhussup/konform/internal/source/common"
)

func TestNewFileSource(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		callerDir string
	}{
		{name: "regular values", path: "config.json", callerDir: "/tmp"},
		{name: "empty values", path: "", callerDir: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFileSource(tt.path, tt.callerDir, common.UnknownKeySuggestionError)
			if got.path != tt.path {
				t.Fatalf("path = %q, want %q", got.path, tt.path)
			}
			if got.callerDir != tt.callerDir {
				t.Fatalf("callerDir = %q, want %q", got.callerDir, tt.callerDir)
			}
			if got.suggestionMode != common.UnknownKeySuggestionError {
				t.Fatalf("suggestionMode = %v, want %v", got.suggestionMode, common.UnknownKeySuggestionError)
			}
		})
	}
}

func TestNewByteSource(t *testing.T) {
	got := NewByteSource([]byte(`{"port":"8080"}`), common.UnknownKeySuggestionError)
	if string(got.data) != `{"port":"8080"}` {
		t.Fatalf("data = %q, want %q", string(got.data), `{"port":"8080"}`)
	}
	if got.suggestionMode != common.UnknownKeySuggestionError {
		t.Fatalf("suggestionMode = %v, want %v", got.suggestionMode, common.UnknownKeySuggestionError)
	}
}

func TestLoadFile(t *testing.T) {
	makeSchema := func(port *int, mode *string) *schema.Schema {
		fields := []schema.Field{{Path: "Port", KeyName: "port", Type: reflect.TypeOf(0), Value: reflect.ValueOf(port).Elem()}}
		if mode != nil {
			fields = append(fields, schema.Field{Path: "Mode", KeyName: "mode", Type: reflect.TypeOf(""), Value: reflect.ValueOf(mode).Elem()})
		}
		return &schema.Schema{Fields: fields}
	}

	tests := []struct {
		name        string
		setup       func(t *testing.T) (Source, *schema.Schema)
		wantErrType error
		wantErrLike []string
		validate    func(t *testing.T, sc *schema.Schema)
	}{
		{name: "nil schema", setup: func(t *testing.T) (Source, *schema.Schema) {
			return NewFileSource("config.json", "", common.UnknownKeySuggestionError), nil
		}, wantErrType: errs.InvalidSchemaNil},
		{name: "missing file", setup: func(t *testing.T) (Source, *schema.Schema) {
			var port int
			return NewFileSource("missing.json", t.TempDir(), common.UnknownKeySuggestionError), makeSchema(&port, nil)
		}, wantErrType: errs.DecodeSourceRead},
		{name: "parse error", setup: func(t *testing.T) (Source, *schema.Schema) {
			dir := t.TempDir()
			p := filepath.Join(dir, "config.json")
			_ = os.WriteFile(p, []byte(`{"port":`), 0o600)
			var port int
			return NewFileSource("config.json", dir, common.UnknownKeySuggestionError), makeSchema(&port, nil)
		}, wantErrType: errs.DecodeSourceParse},
		{name: "decode field error", setup: func(t *testing.T) (Source, *schema.Schema) {
			dir := t.TempDir()
			p := filepath.Join(dir, "config.json")
			_ = os.WriteFile(p, []byte(`{"port":"8080","mode":true}`), 0o600)
			var port int
			var mode string
			return NewFileSource("config.json", dir, common.UnknownKeySuggestionError), makeSchema(&port, &mode)
		}, wantErrType: errs.DecodeSourceField, wantErrLike: []string{`json "mode" -> Mode`, "expected string, got bool"}},
		{name: "success", setup: func(t *testing.T) (Source, *schema.Schema) {
			dir := t.TempDir()
			p := filepath.Join(dir, "config.json")
			_ = os.WriteFile(p, []byte(`{"port":"8080"}`), 0o600)
			var port int
			return NewFileSource("config.json", dir, common.UnknownKeySuggestionError), makeSchema(&port, nil)
		}, validate: func(t *testing.T, sc *schema.Schema) {
			if got := sc.Fields[0].Value.Interface().(int); got != 8080 {
				t.Fatalf("Port = %d, want 8080", got)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, sc := tt.setup(t)
			err := source.LoadFile(sc)
			if tt.wantErrType != nil {
				if err == nil || !errors.Is(err, tt.wantErrType) {
					t.Fatalf("LoadFile() error = %v, want wrapped %v", err, tt.wantErrType)
				}
				for _, part := range tt.wantErrLike {
					if !strings.Contains(err.Error(), part) {
						t.Fatalf("LoadFile() error = %q, want to contain %q", err.Error(), part)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadFile() error = %v, want nil", err)
			}
			if tt.validate != nil {
				tt.validate(t, sc)
			}
		})
	}
}

func TestLoadBytes(t *testing.T) {
	makeSchema := func(port *int, mode *string) *schema.Schema {
		fields := []schema.Field{{Path: "Port", KeyName: "port", Type: reflect.TypeOf(0), Value: reflect.ValueOf(port).Elem()}}
		if mode != nil {
			fields = append(fields, schema.Field{Path: "Mode", KeyName: "mode", Type: reflect.TypeOf(""), Value: reflect.ValueOf(mode).Elem()})
		}
		return &schema.Schema{Fields: fields}
	}

	tests := []struct {
		name        string
		setup       func() (Source, *schema.Schema)
		wantErrType error
		wantErrLike []string
		validate    func(t *testing.T, sc *schema.Schema)
	}{
		{name: "nil schema", setup: func() (Source, *schema.Schema) {
			return NewByteSource([]byte(`{"port":"8080"}`), common.UnknownKeySuggestionError), nil
		}, wantErrType: errs.InvalidSchemaNil},
		{name: "parse error", setup: func() (Source, *schema.Schema) {
			var port int
			return NewByteSource([]byte(`{"port":`), common.UnknownKeySuggestionError), makeSchema(&port, nil)
		}, wantErrType: errs.DecodeSourceParse},
		{name: "decode field error", setup: func() (Source, *schema.Schema) {
			var port int
			var mode string
			return NewByteSource([]byte(`{"port":"8080","mode":true}`), common.UnknownKeySuggestionError), makeSchema(&port, &mode)
		}, wantErrType: errs.DecodeSourceField, wantErrLike: []string{`json "mode" -> Mode`, "expected string, got bool"}},
		{name: "success", setup: func() (Source, *schema.Schema) {
			var port int
			return NewByteSource([]byte(`{"port":"8080"}`), common.UnknownKeySuggestionError), makeSchema(&port, nil)
		}, validate: func(t *testing.T, sc *schema.Schema) {
			if got := sc.Fields[0].Value.Interface().(int); got != 8080 {
				t.Fatalf("Port = %d, want 8080", got)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, sc := tt.setup()
			err := source.LoadBytes(sc)
			if tt.wantErrType != nil {
				if err == nil || !errors.Is(err, tt.wantErrType) {
					t.Fatalf("LoadBytes() error = %v, want wrapped %v", err, tt.wantErrType)
				}
				for _, part := range tt.wantErrLike {
					if !strings.Contains(err.Error(), part) {
						t.Fatalf("LoadBytes() error = %q, want to contain %q", err.Error(), part)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadBytes() error = %v, want nil", err)
			}
			if tt.validate != nil {
				tt.validate(t, sc)
			}
		})
	}
}
