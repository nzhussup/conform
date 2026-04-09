package common

import (
	"fmt"
	"os"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

type Document map[string]any

type UnmarshalFunc func([]byte) (Document, error)

func LoadFileWithMode(sc *schema.Schema, path string, callerDir string, format string, unmarshal UnmarshalFunc, suggestionMode UnknownKeySuggestionMode) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	data, err := os.ReadFile(resolvePath(path, callerDir))
	if err != nil {
		return errs.WrapDecode(errs.DecodeSourceRead, fmt.Sprintf("%s file %q", format, path), err)
	}

	return decodeAndApplyWithMode(
		sc,
		data,
		format,
		unmarshal,
		suggestionMode,
		defaultSourceLabel(path, format),
		fmt.Sprintf("%s file", format),
	)
}

func LoadBytesWithMode(sc *schema.Schema, data []byte, format string, unmarshal UnmarshalFunc, suggestionMode UnknownKeySuggestionMode, sourceLabel string) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	return decodeAndApplyWithMode(sc, data, format, unmarshal, suggestionMode, sourceLabel, format)
}

func decodeAndApplyWithMode(
	sc *schema.Schema,
	data []byte,
	format string,
	unmarshal UnmarshalFunc,
	suggestionMode UnknownKeySuggestionMode,
	sourceLabel string,
	parseContext string,
) error {
	doc, err := unmarshal(data)
	if err != nil {
		return errs.WrapDecode(errs.DecodeSourceParse, parseContext, err)
	}

	return ApplyWithMode(sc, doc, format, suggestionMode, sourceLabel)
}
