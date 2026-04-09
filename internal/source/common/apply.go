package common

import (
	"errors"
	"fmt"
	"os"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func Apply(sc *schema.Schema, doc Document, format string) error {
	return ApplyWithMode(sc, doc, format, UnknownKeySuggestionWarn, format)
}

func ApplyWithMode(sc *schema.Schema, doc Document, format string, suggestionMode UnknownKeySuggestionMode, sourceLabel string) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	pathAliases := BuildPathAliases(sc)
	fieldErrors := make([]error, 0, len(sc.Fields))
	fieldErrors = append(fieldErrors, collectUnknownKeyErrors(sc, doc, pathAliases, format, suggestionMode)...)
	fieldErrors = append(fieldErrors, applyDocumentFields(sc, doc, pathAliases, format, sourceLabel)...)

	if len(fieldErrors) > 0 {
		return errors.Join(fieldErrors...)
	}
	return nil
}

func collectUnknownKeyErrors(
	sc *schema.Schema,
	doc Document,
	pathAliases map[string]string,
	format string,
	suggestionMode UnknownKeySuggestionMode,
) []error {
	issues := FindUnknownKeyIssuesWithMode(sc, doc, pathAliases, suggestionMode)
	if len(issues) == 0 {
		return nil
	}

	fieldErrors := make([]error, 0, len(issues))
	for _, issue := range issues {
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

	return fieldErrors
}

func applyDocumentFields(
	sc *schema.Schema,
	doc Document,
	pathAliases map[string]string,
	format string,
	sourceLabel string,
) []error {
	fieldErrors := make([]error, 0, len(sc.Fields))

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

	return fieldErrors
}
