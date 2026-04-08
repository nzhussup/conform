package common

import "testing"

func TestUnknownKeySuggestionModeConstants(t *testing.T) {
	if UnknownKeySuggestionWarn != 0 {
		t.Fatalf("UnknownKeySuggestionWarn = %d, want 0", UnknownKeySuggestionWarn)
	}
	if UnknownKeySuggestionError != 1 {
		t.Fatalf("UnknownKeySuggestionError = %d, want 1", UnknownKeySuggestionError)
	}
	if UnknownKeySuggestionOff != 2 {
		t.Fatalf("UnknownKeySuggestionOff = %d, want 2", UnknownKeySuggestionOff)
	}
	if UnknownKeySuggestionWarn == UnknownKeySuggestionError {
		t.Fatalf("suggestion modes must be distinct")
	}
	if UnknownKeySuggestionWarn == UnknownKeySuggestionOff || UnknownKeySuggestionError == UnknownKeySuggestionOff {
		t.Fatalf("all suggestion modes must be distinct")
	}
}
