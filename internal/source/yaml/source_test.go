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
	"github.com/nzhussup/konform/internal/source/common"
)

func TestNewFileSource(t *testing.T) {
	got := NewFileSource("config.yaml", "/tmp", common.UnknownKeySuggestionError)
	if got.path != "config.yaml" || got.callerDir != "/tmp" {
		t.Fatalf("unexpected source: %#v", got)
	}
}

func TestNewByteSource(t *testing.T) {
	got := NewByteSource([]byte("port: \"8080\"\n"), common.UnknownKeySuggestionError)
	if string(got.data) != "port: \"8080\"\n" {
		t.Fatalf("data = %q", string(got.data))
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
	}{
		{name: "nil schema", setup: func(t *testing.T) (Source, *schema.Schema) {
			return NewFileSource("config.yaml", "", common.UnknownKeySuggestionError), nil
		}, wantErrType: errs.InvalidSchemaNil},
		{name: "missing file", setup: func(t *testing.T) (Source, *schema.Schema) {
			var port int
			return NewFileSource("missing.yaml", t.TempDir(), common.UnknownKeySuggestionError), makeSchema(&port, nil)
		}, wantErrType: errs.DecodeSourceRead},
		{name: "parse error", setup: func(t *testing.T) (Source, *schema.Schema) {
			dir := t.TempDir()
			p := filepath.Join(dir, "config.yaml")
			_ = os.WriteFile(p, []byte("port: ["), 0o600)
			var port int
			return NewFileSource("config.yaml", dir, common.UnknownKeySuggestionError), makeSchema(&port, nil)
		}, wantErrType: errs.DecodeSourceParse},
		{name: "decode field error", setup: func(t *testing.T) (Source, *schema.Schema) {
			dir := t.TempDir()
			p := filepath.Join(dir, "config.yaml")
			_ = os.WriteFile(p, []byte("port: \"8080\"\nmode: true\n"), 0o600)
			var port int
			var mode string
			return NewFileSource("config.yaml", dir, common.UnknownKeySuggestionError), makeSchema(&port, &mode)
		}, wantErrType: errs.DecodeSourceField, wantErrLike: []string{`yaml "mode" -> Mode`, "expected string, got bool"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, sc := tt.setup(t)
			err := source.LoadFile(sc)
			if tt.wantErrType == nil {
				if err != nil {
					t.Fatalf("LoadFile() error = %v", err)
				}
				return
			}
			if err == nil || !errors.Is(err, tt.wantErrType) {
				t.Fatalf("LoadFile() error = %v, want %v", err, tt.wantErrType)
			}
			for _, part := range tt.wantErrLike {
				if !strings.Contains(err.Error(), part) {
					t.Fatalf("LoadFile() error = %q, want %q", err.Error(), part)
				}
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

	var port int
	var mode string
	source := NewByteSource([]byte("port: \"8080\"\nmode: true\n"), common.UnknownKeySuggestionError)
	sc := makeSchema(&port, &mode)
	err := source.LoadBytes(sc)
	if err == nil || !errors.Is(err, errs.DecodeSourceField) {
		t.Fatalf("LoadBytes() error = %v, want %v", err, errs.DecodeSourceField)
	}

	port = 0
	source = NewByteSource([]byte("port: \"8080\"\n"), common.UnknownKeySuggestionError)
	sc = makeSchema(&port, nil)
	if err := source.LoadBytes(sc); err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if port != 8080 {
		t.Fatalf("Port = %d, want 8080", port)
	}
}
