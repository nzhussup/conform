package common

import (
	"reflect"
	"testing"

	"github.com/nzhussup/konform/internal/schema"
)

func TestFindUnknownKeyIssues(t *testing.T) {
	type nested struct{ Port int }

	makeSchema := func() *schema.Schema {
		var s nested
		var port int
		return &schema.Schema{
			Fields: []schema.Field{
				{
					Path:    "Server",
					KeyName: "server_cfg",
					Type:    reflect.TypeOf(s),
					Value:   reflect.ValueOf(&s).Elem(),
				},
				{
					Path:  "Server.Port",
					Type:  reflect.TypeOf(0),
					Value: reflect.ValueOf(&port).Elem(),
				},
			},
		}
	}

	t.Run("nil schema", func(t *testing.T) {
		got := FindUnknownKeyIssues(nil, Document{"server_cfg": map[string]any{"Port": 8080}}, nil)
		if got != nil {
			t.Fatalf("FindUnknownKeyIssues() = %#v, want nil", got)
		}
	})

	t.Run("unexpected input path with suggestion", func(t *testing.T) {
		sc := makeSchema()
		aliases := BuildPathAliases(sc)
		doc := Document{
			"server_cfg": map[string]any{
				"Porrt": 8080,
			},
		}

		got := FindUnknownKeyIssues(sc, doc, aliases)
		if len(got) != 1 {
			t.Fatalf("len(issues) = %d, want 1", len(got))
		}
		if got[0].Path != "server_cfg.Porrt" {
			t.Fatalf("issue path = %q, want %q", got[0].Path, "server_cfg.Porrt")
		}
		if got[0].Suggestion != "server_cfg.Port" {
			t.Fatalf("issue suggestion = %q, want %q", got[0].Suggestion, "server_cfg.Port")
		}
	})

	t.Run("present expected path has no issues", func(t *testing.T) {
		sc := makeSchema()
		aliases := BuildPathAliases(sc)
		doc := Document{
			"server_cfg": map[string]any{
				"Port": 8080,
			},
		}

		got := FindUnknownKeyIssues(sc, doc, aliases)
		if len(got) != 0 {
			t.Fatalf("len(issues) = %d, want 0", len(got))
		}
	})
}

func TestFindUnknownKeyIssuesWithMode(t *testing.T) {
	var appName string
	sc := &schema.Schema{
		Fields: []schema.Field{
			{
				Path:  "AppName",
				Type:  reflect.TypeOf(""),
				Value: reflect.ValueOf(&appName).Elem(),
			},
		},
	}
	aliases := BuildPathAliases(sc)
	doc := Document{
		"App": map[string]any{
			"Name": "konform",
		},
	}

	t.Run("error mode reports unexpected key with schema suggestion", func(t *testing.T) {
		issues := FindUnknownKeyIssuesWithMode(sc, doc, aliases, UnknownKeySuggestionError)
		if len(issues) != 1 {
			t.Fatalf("len(issues) = %d, want 1", len(issues))
		}
		if issues[0].Path != "App.Name" || issues[0].Suggestion != "AppName" {
			t.Fatalf("issue = %#v, want path App.Name with suggestion AppName", issues[0])
		}
	})

	t.Run("off mode reports no issues", func(t *testing.T) {
		issues := FindUnknownKeyIssuesWithMode(sc, doc, aliases, UnknownKeySuggestionOff)
		if len(issues) != 0 {
			t.Fatalf("len(issues) = %d, want 0", len(issues))
		}
	})

	t.Run("warn mode reports unknown actual file key", func(t *testing.T) {
		issues := FindUnknownKeyIssuesWithMode(sc, doc, aliases, UnknownKeySuggestionWarn)
		if len(issues) != 1 {
			t.Fatalf("len(issues) = %d, want 1", len(issues))
		}
		if issues[0].Path != "App.Name" {
			t.Fatalf("issue path = %q, want %q", issues[0].Path, "App.Name")
		}
		if issues[0].Suggestion != "AppName" {
			t.Fatalf("issue suggestion = %q, want %q", issues[0].Suggestion, "AppName")
		}
	})
}

func TestUnknownKeysBuildExpectedLookupPaths(t *testing.T) {
	type nested struct{ Port int }
	var s nested
	var port int
	var timeout int

	sc := &schema.Schema{
		Fields: []schema.Field{
			{
				Path:    "Server",
				KeyName: "server_cfg",
				Type:    reflect.TypeOf(s),
				Value:   reflect.ValueOf(&s).Elem(),
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

func TestUnknownKeysBuildExpectedLookupPathsSkipsEmptyLookup(t *testing.T) {
	var value string
	sc := &schema.Schema{
		Fields: []schema.Field{
			{
				Path:  "",
				Type:  reflect.TypeOf(""),
				Value: reflect.ValueOf(&value).Elem(),
			},
		},
	}

	got := BuildExpectedLookupPaths(sc, map[string]string{})
	if len(got) != 0 {
		t.Fatalf("BuildExpectedLookupPaths() = %#v, want empty", got)
	}
}

func TestUnknownKeysFindUnknownKeys(t *testing.T) {
	doc := Document{
		"server": map[string]any{
			"port": 8080,
			"host": "localhost",
		},
		"mode": "prod",
	}
	expected := map[string]struct{}{
		"server.port": {},
	}

	got := FindUnknownKeys(doc, expected)
	want := []string{"mode", "server.host"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FindUnknownKeys() = %#v, want %#v", got, want)
	}
}

func TestUnknownKeysFlattenLeafPathsAndSliceToPathSet(t *testing.T) {
	doc := Document{
		"server": map[string]any{
			"port": 8080,
			"tls": map[string]any{
				"enabled": true,
			},
		},
		"empty": map[string]any{},
	}

	paths := FlattenLeafPaths(doc)
	wantPaths := []string{"empty", "server.port", "server.tls.enabled"}
	if !reflect.DeepEqual(paths, wantPaths) {
		t.Fatalf("FlattenLeafPaths() = %#v, want %#v", paths, wantPaths)
	}

	set := sliceToPathSet(paths)
	for _, p := range wantPaths {
		if _, ok := set[p]; !ok {
			t.Fatalf("sliceToPathSet() missing path %q", p)
		}
	}
}

func TestUnknownKeysFlattenLeafPathsNilDoc(t *testing.T) {
	if got := FlattenLeafPaths(nil); got != nil {
		t.Fatalf("FlattenLeafPaths(nil) = %#v, want nil", got)
	}
}

func TestFlattenLeafPathsSkipsEmptyPrefixForNonMapRoot(t *testing.T) {
	paths := make([]string, 0)
	flattenLeafPaths("", 42, &paths)
	if len(paths) != 0 {
		t.Fatalf("flattenLeafPaths('', non-map) = %#v, want empty", paths)
	}
}
