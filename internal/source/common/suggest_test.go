package common

import "testing"

func TestSuggestPathUnit(t *testing.T) {
	tests := []struct {
		name        string
		unknownPath string
		candidates  map[string]struct{}
		wantPath    string
		wantOK      bool
	}{
		{
			name:        "empty unknown path",
			unknownPath: "",
			candidates: map[string]struct{}{
				"server.port": {},
			},
			wantOK: false,
		},
		{
			name:        "empty candidate key returns no suggestion",
			unknownPath: "server.port",
			candidates: map[string]struct{}{
				"": {},
			},
			wantOK: false,
		},
		{
			name:        "empty candidates",
			unknownPath: "server.port",
			candidates:  map[string]struct{}{},
			wantOK:      false,
		},
		{
			name:        "closest suggestion is returned",
			unknownPath: "server.poort",
			candidates: map[string]struct{}{
				"server.port": {},
				"server.host": {},
			},
			wantPath: "server.port",
			wantOK:   true,
		},
		{
			name:        "too far candidate is rejected",
			unknownPath: "a",
			candidates: map[string]struct{}{
				"zzzz": {},
			},
			wantOK: false,
		},
		{
			name:        "max distance cap still applies for long strings",
			unknownPath: "this.is.a.very.long.unknown.path",
			candidates: map[string]struct{}{
				"x": {},
			},
			wantOK: false,
		},
		{
			name:        "ties choose lexicographically smaller path",
			unknownPath: "server.pxrt",
			candidates: map[string]struct{}{
				"server.part": {},
				"server.port": {},
			},
			wantPath: "server.part",
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotOK := SuggestPath(tt.unknownPath, tt.candidates)
			if gotOK != tt.wantOK {
				t.Fatalf("SuggestPath() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("SuggestPath() path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  int
	}{
		{name: "equal strings", left: "port", right: "port", want: 0},
		{name: "single substitution", left: "poort", right: "port", want: 1},
		{name: "single insertion", left: "port", right: "ports", want: 1},
		{name: "single deletion", left: "ports", right: "port", want: 1},
		{name: "unicode runes", left: "päth", right: "path", want: 1},
		{name: "left empty", left: "", right: "path", want: 4},
		{name: "right empty", left: "path", right: "", want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := levenshteinDistance(tt.left, tt.right)
			if got != tt.want {
				t.Fatalf("levenshteinDistance() = %d, want %d", got, tt.want)
			}
		})
	}
}
