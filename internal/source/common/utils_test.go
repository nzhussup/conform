package common

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/nzhussup/konform/internal/schema"
)

func TestResolvePath(t *testing.T) {
	t.Run("absolute path is preserved", func(t *testing.T) {
		abs := filepath.Join(string(filepath.Separator), "tmp", "config.yaml")
		if got := resolvePath(abs, "/caller"); got != abs {
			t.Fatalf("resolvePath() = %q, want %q", got, abs)
		}
	})

	t.Run("relative path uses caller dir", func(t *testing.T) {
		got := resolvePath("config.yaml", "/caller")
		want := filepath.Join("/caller", "config.yaml")
		if got != want {
			t.Fatalf("resolvePath() = %q, want %q", got, want)
		}
	})

	t.Run("empty caller dir leaves path unchanged", func(t *testing.T) {
		if got := resolvePath("config.yaml", ""); got != "config.yaml" {
			t.Fatalf("resolvePath() = %q, want %q", got, "config.yaml")
		}
	})
}

func TestDefaultSourceLabel(t *testing.T) {
	t.Run("path wins when present", func(t *testing.T) {
		if got := defaultSourceLabel("config.yaml", "yaml"); got != "config.yaml" {
			t.Fatalf("defaultSourceLabel() = %q, want %q", got, "config.yaml")
		}
	})

	t.Run("format used when path empty", func(t *testing.T) {
		if got := defaultSourceLabel("", "yaml"); got != "yaml" {
			t.Fatalf("defaultSourceLabel() = %q, want %q", got, "yaml")
		}
	})
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
		{
			name: "full path alias replaces whole path",
			field: schema.Field{
				Path: "Server.Port",
			},
			pathAliases: map[string]string{
				"Server.Port": "server_port",
			},
			want: "server_port",
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
