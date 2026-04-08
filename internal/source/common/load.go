package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/nzhussup/konform/internal/decode"
	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

type Document map[string]any

type UnmarshalFunc func([]byte) (Document, error)

func LoadFile(sc *schema.Schema, path string, callerDir string, format string, unmarshal UnmarshalFunc) error {
	return LoadFileWithMode(sc, path, callerDir, format, unmarshal, UnknownKeySuggestionWarn)
}

func LoadFileWithMode(sc *schema.Schema, path string, callerDir string, format string, unmarshal UnmarshalFunc, suggestionMode UnknownKeySuggestionMode) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	resolvedPath := resolvePath(path, callerDir)

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return errs.WrapDecode(errs.DecodeSourceRead, fmt.Sprintf("%s file %q", format, path), err)
	}

	doc, err := unmarshal(data)
	if err != nil {
		return errs.WrapDecode(errs.DecodeSourceParse, fmt.Sprintf("%s file", format), err)
	}

	sourceLabel := path
	if sourceLabel == "" {
		sourceLabel = format
	}

	return ApplyWithMode(sc, doc, format, suggestionMode, sourceLabel)
}

func resolvePath(path string, callerDir string) string {
	if filepath.IsAbs(path) || callerDir == "" {
		return path
	}
	return filepath.Join(callerDir, path)
}

func Apply(sc *schema.Schema, doc Document, format string) error {
	return ApplyWithMode(sc, doc, format, UnknownKeySuggestionWarn, format)
}

func ApplyWithMode(sc *schema.Schema, doc Document, format string, suggestionMode UnknownKeySuggestionMode, sourceLabel string) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	pathAliases := BuildPathAliases(sc)
	fieldErrors := make([]error, 0, len(sc.Fields))

	unknownKeyIssues := FindUnknownKeyIssuesWithMode(sc, doc, pathAliases, suggestionMode)
	for _, issue := range unknownKeyIssues {
		msg := fmt.Sprintf(`unknown configuration key %q`, issue.Path)
		if issue.Suggestion != "" {
			msg = fmt.Sprintf(`%s (did you mean %q?)`, msg, issue.Suggestion)
		}
		if suggestionMode == UnknownKeySuggestionWarn {
			fmt.Fprintf(os.Stderr, "konform: warning: %s: %s\n", format, msg)
			continue
		}
		fieldErrors = append(fieldErrors, errs.WrapDecode(errs.DecodeSourceField, format, errors.New(msg)))
	}

	for i := range sc.Fields {
		field := sc.Fields[i]
		if isStructField(field) {
			continue
		}

		lookupPath := ResolveLookupPath(field, pathAliases)
		value, ok := GetByPath(doc, lookupPath)
		if !ok {
			continue
		}

		if err := setFieldFromValue(field, value); err != nil {
			ctx := fmt.Sprintf("%s %q -> %s", format, lookupPath, field.Path)
			fieldErrors = append(fieldErrors, errs.WrapDecode(errs.DecodeSourceField, ctx, err))
			continue
		}
		sc.Fields[i].Source = sourceLabel
	}

	if len(fieldErrors) > 0 {
		return errors.Join(fieldErrors...)
	}

	return nil
}

func isStructField(field schema.Field) bool {
	return field.Type.Kind() == reflect.Struct
}

func setFieldFromValue(field schema.Field, value any) error {
	return decode.SetFieldValue(field, value)
}

func BuildPathAliases(sc *schema.Schema) map[string]string {
	aliases := make(map[string]string, len(sc.Fields))
	for _, field := range sc.Fields {
		if field.KeyName == "" {
			continue
		}
		aliases[field.Path] = field.KeyName
	}
	return aliases
}

func ResolveLookupPath(field schema.Field, pathAliases map[string]string) string {
	if field.KeyName != "" {
		return field.KeyName
	}

	resolved := field.Path
	parts := strings.Split(field.Path, ".")

	for i := range parts {
		prefix := joinPath(parts[:i+1])
		alias, ok := pathAliases[prefix]
		if !ok {
			continue
		}

		suffix := joinPath(parts[i+1:])
		if suffix == "" {
			resolved = alias
		} else {
			resolved = alias + "." + suffix
		}
	}

	return resolved
}

func joinPath(parts []string) string {
	return strings.Join(parts, ".")
}

func GetByPath(doc Document, path string) (any, bool) {
	if path == "" {
		return nil, false
	}

	keys := strings.Split(path, ".")
	current := any(doc)

	for _, key := range keys {
		m, ok := asStringMap(current)
		if !ok {
			return nil, false
		}
		current, ok = m[key]
		if !ok {
			return nil, false
		}
	}

	return current, true
}

func asStringMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case Document:
		return map[string]any(m), true
	case map[string]any:
		return m, true
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, val := range m {
			key, ok := k.(string)
			if !ok {
				return nil, false
			}
			out[key] = val
		}
		return out, true
	default:
		return nil, false
	}
}
