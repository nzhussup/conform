package common

type UnknownKeySuggestionMode int

const (
	UnknownKeySuggestionWarn UnknownKeySuggestionMode = iota
	UnknownKeySuggestionError
	UnknownKeySuggestionOff
)
