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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.args
			if tt.setup != nil {
				a = tt.setup(t)
			}

			err := LoadFile(a.sc, a.path, a.callerDir, a.format, a.unmarshal)
			if tt.wantErrType == nil {
				if err != nil {
					t.Fatalf("LoadFile() error = %v, want nil", err)
				}
				if tt.validate != nil {
					tt.validate(t, a)
				}
				return
			}

			if err == nil {
				t.Fatalf("LoadFile() error = nil, want %v", tt.wantErrType)
			}
			if !errors.Is(err, tt.wantErrType) {
				t.Fatalf("LoadFile() error = %v, want wrapped %v", err, tt.wantErrType)
			}
			for _, part := range tt.wantErrLike {
				if !strings.Contains(err.Error(), part) {
					t.Fatalf("LoadFile() error = %q, want to contain %q", err.Error(), part)
				}
			}
		})
	}
}

func TestApply(t *testing.T) {
	type nested struct{}

	tests := []struct {
		name        string
		scBuilder   func() *schema.Schema
		doc         Document
		mode        UnknownKeySuggestionMode
		wantErrType error
		wantErrLike []string
		validate    func(t *testing.T, sc *schema.Schema)
	}{
		{
			name:        "nil schema",
			scBuilder:   func() *schema.Schema { return nil },
			doc:         Document{},
			mode:        UnknownKeySuggestionOff,
			wantErrType: errs.InvalidSchemaNil,
		},
		{
			name: "missing path is ignored",
			scBuilder: func() *schema.Schema {
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Port",
							KeyName: "port",
							Type:    reflect.TypeOf(0),
							Value:   reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionOff,
			doc:  Document{},
			validate: func(t *testing.T, sc *schema.Schema) {
				t.Helper()
				if got := sc.Fields[0].Value.Interface().(int); got != 0 {
					t.Fatalf("port = %d, want 0", got)
				}
			},
		},
		{
			name: "decode error is wrapped with field context",
			scBuilder: func() *schema.Schema {
				var enabled string
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Enabled",
							KeyName: "enabled",
							Type:    reflect.TypeOf(""),
							Value:   reflect.ValueOf(&enabled).Elem(),
						},
					},
				}
			},
			mode:        UnknownKeySuggestionOff,
			doc:         Document{"enabled": true},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{"yaml", "enabled", "Enabled", "expected string, got bool"},
		},
		{
			name: "collects multiple decode errors",
			scBuilder: func() *schema.Schema {
				var name string
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Name",
							KeyName: "name",
							Type:    reflect.TypeOf(""),
							Value:   reflect.ValueOf(&name).Elem(),
						},
						{
							Path:    "Port",
							KeyName: "port",
							Type:    reflect.TypeOf(0),
							Value:   reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode:        UnknownKeySuggestionOff,
			doc:         Document{"name": true, "port": "not-int"},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{"yaml \"name\" -> Name", "expected string, got bool", "yaml \"port\" -> Port", "invalid int value"},
		},
		{
			name: "nested field uses parent alias",
			scBuilder: func() *schema.Schema {
				var parent nested
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Server",
							KeyName: "server_cfg",
							Type:    reflect.TypeOf(parent),
							Value:   reflect.ValueOf(&parent).Elem(),
						},
						{
							Path:  "Server.Port",
							Type:  reflect.TypeOf(0),
							Value: reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionOff,
			doc: Document{
				"server_cfg": map[string]any{
					"Port": "9090",
				},
			},
			validate: func(t *testing.T, sc *schema.Schema) {
				t.Helper()
				if got := sc.Fields[1].Value.Interface().(int); got != 9090 {
					t.Fatalf("port = %d, want 9090", got)
				}
			},
		},
		{
			name: "explicit key typo reports unexpected config key with schema suggestion",
			scBuilder: func() *schema.Schema {
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "AppName",
							KeyName: "App.Nam",
							Type:    reflect.TypeOf(0),
							Value:   reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionError,
			doc: Document{
				"App": map[string]any{
					"Name": 8080,
				},
			},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{`unknown configuration key "App.Name"`, `did you mean "App.Nam"?`},
		},
		{
			name: "non explicit path typo reports unexpected config key with schema suggestion",
			scBuilder: func() *schema.Schema {
				type server struct{ Port int }
				var s server
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:  "Server",
							Type:  reflect.TypeOf(s),
							Value: reflect.ValueOf(&s).Elem(),
						},
						{
							Path:  "Server.Port",
							Type:  reflect.TypeOf(0),
							Value: reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionError,
			doc: Document{
				"Server": map[string]any{
					"Porrt": 8080,
				},
			},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{`unknown configuration key "Server.Porrt"`, `did you mean "Server.Port"?`},
		},
		{
			name: "unexpected dotted key suggests struct key",
			scBuilder: func() *schema.Schema {
				var appName string
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:  "AppName",
							Type:  reflect.TypeOf(""),
							Value: reflect.ValueOf(&appName).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionError,
			doc: Document{
				"App": map[string]any{
					"Name": "konform",
				},
			},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{`unknown configuration key "App.Name"`, `did you mean "AppName"?`},
		},
		{
			name: "extra file keys are ignored in non-strict mode",
			scBuilder: func() *schema.Schema {
				var appDebug bool
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:  "AppDebug",
							Type:  reflect.TypeOf(true),
							Value: reflect.ValueOf(&appDebug).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionOff,
			doc: Document{
				"AppDebug": true,
				"App": map[string]any{
					"Debug": true,
				},
			},
			validate: func(t *testing.T, sc *schema.Schema) {
				t.Helper()
				if got := sc.Fields[0].Value.Interface().(bool); !got {
					t.Fatalf("AppDebug = %v, want true", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := tt.scBuilder()
			err := ApplyWithMode(sc, tt.doc, "yaml", tt.mode, "config.yaml")

			if tt.wantErrType == nil {
				if err != nil {
					t.Fatalf("Apply() error = %v, want nil", err)
				}
				if tt.validate != nil {
					tt.validate(t, sc)
				}
				return
			}

			if err == nil {
				t.Fatalf("Apply() error = nil, want %v", tt.wantErrType)
			}
			if !errors.Is(err, tt.wantErrType) {
				t.Fatalf("Apply() error = %v, want wrapped %v", err, tt.wantErrType)
			}
			for _, part := range tt.wantErrLike {
				if !strings.Contains(err.Error(), part) {
					t.Fatalf("Apply() error = %q, want to contain %q", err.Error(), part)
				}
			}
		})
	}
}

func TestBuildExpectedLookupPaths(t *testing.T) {
	type nested struct{}

	var parent nested
	var port int
	var timeout int

	sc := &schema.Schema{
		Fields: []schema.Field{
			{
				Path:    "Server",
				KeyName: "server_cfg",
				Type:    reflect.TypeOf(parent),
				Value:   reflect.ValueOf(&parent).Elem(),
			},
			{
				Path:  "Server.Port",
				Type:  reflect.TypeOf(0),
				Value: reflect.ValueOf(&port).Elem(),
			},
			{
				Path:    "Timeout",
				KeyName: "timeout",
				Type:    reflect.TypeOf(0),
				Value:   reflect.ValueOf(&timeout).Elem(),
			},
		},
	}

	got := BuildExpectedLookupPaths(sc, BuildPathAliases(sc))
	want := map[string]struct{}{
		"server_cfg.Port": {},
		"timeout":         {},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildExpectedLookupPaths() = %#v, want %#v", got, want)
	}
}

func TestFlattenLeafPaths(t *testing.T) {
	doc := Document{
		"server": map[string]any{
			"port": 8080,
			"tls": map[string]any{
				"enabled": true,
			},
		},
		"mode": "prod",
	}

	got := FlattenLeafPaths(doc)
	want := []string{"mode", "server.port", "server.tls.enabled"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FlattenLeafPaths() = %#v, want %#v", got, want)
	}
}

func TestSuggestPath(t *testing.T) {
	expected := map[string]struct{}{
		"server.port": {},
		"server.host": {},
		"log.level":   {},
	}

	got, ok := SuggestPath("server.poort", expected)
	if !ok {
		t.Fatalf("SuggestPath() expected suggestion, got none")
	}
	if got != "server.port" {
		t.Fatalf("SuggestPath() = %q, want %q", got, "server.port")
	}

	got, ok = SuggestPath("completely.different.key", expected)
	if ok {
		t.Fatalf("SuggestPath() expected no suggestion, got %q", got)
	}
}

func TestBuildPathAliases(t *testing.T) {
	tests := []struct {
		name string
		sc   *schema.Schema
		want map[string]string
	}{
		{
			name: "collects only fields with key names",
			sc: &schema.Schema{
				Fields: []schema.Field{
					{Path: "Server", KeyName: "server_cfg"},
					{Path: "Server.Port", KeyName: ""},
					{Path: "DB.Host", KeyName: "db_host"},
				},
			},
			want: map[string]string{
				"Server":  "server_cfg",
				"DB.Host": "db_host",
			},
		},
		{
			name: "empty schema",
			sc:   &schema.Schema{},
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildPathAliases(tt.sc)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("BuildPathAliases() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestResolveLookupPath(t *testing.T) {
	tests := []struct {
		name        string
		field       schema.Field
		pathAliases map[string]string
		want        string
	}{
		{
			name: "uses field key name directly",
			field: schema.Field{
				Path:    "Server.Port",
				KeyName: "server_port",
			},
			pathAliases: map[string]string{
				"Server": "server_cfg",
			},
			want: "server_port",
		},
		{
			name: "uses parent alias for nested field",
			field: schema.Field{
				Path: "Server.Port",
			},
			pathAliases: map[string]string{
				"Server": "server_cfg",
			},
			want: "server_cfg.Port",
		},
		{
			name: "prefers deepest alias when multiple match",
			field: schema.Field{
				Path: "Server.DB.Port",
			},
			pathAliases: map[string]string{
				"Server":    "server_cfg",
				"Server.DB": "db_cfg",
			},
			want: "db_cfg.Port",
		},
		{
			name: "no alias returns original path",
			field: schema.Field{
				Path: "Server.Port",
			},
			pathAliases: map[string]string{},
			want:        "Server.Port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLookupPath(tt.field, tt.pathAliases)
			if got != tt.want {
				t.Fatalf("ResolveLookupPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetByPath(t *testing.T) {
	tests := []struct {
		name      string
		doc       Document
		path      string
		want      any
		wantFound bool
	}{
		{
			name:      "empty path",
			doc:       Document{"a": 1},
			path:      "",
			want:      nil,
			wantFound: false,
		},
		{
			name:      "top level key",
			doc:       Document{"a": 1},
			path:      "a",
			want:      1,
			wantFound: true,
		},
		{
			name: "nested map string-any",
			doc: Document{
				"server": map[string]any{
					"port": 8080,
				},
			},
			path:      "server.port",
			want:      8080,
			wantFound: true,
		},
		{
			name: "nested Document map",
			doc: Document{
				"server": Document{
					"host": "localhost",
				},
			},
			path:      "server.host",
			want:      "localhost",
			wantFound: true,
		},
		{
			name: "nested map interface-any with string keys",
			doc: Document{
				"server": map[any]any{
					"port": 9000,
				},
			},
			path:      "server.port",
			want:      9000,
			wantFound: true,
		},
		{
			name: "missing nested key",
			doc: Document{
				"server": map[string]any{},
			},
			path:      "server.port",
			want:      nil,
			wantFound: false,
		},
		{
			name: "non-map encountered in middle",
			doc: Document{
				"server": "localhost",
			},
			path:      "server.port",
			want:      nil,
			wantFound: false,
		},
		{
			name: "map interface-any with non-string key fails",
			doc: Document{
				"server": map[any]any{
					1: "bad",
				},
			},
			path:      "server.port",
			want:      nil,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetByPath(tt.doc, tt.path)
			if ok != tt.wantFound {
				t.Fatalf("GetByPath() found = %v, want %v", ok, tt.wantFound)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetByPath() value = %#v, want %#v", got, tt.want)
			}
		})
	}
}
