package konform

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	internalschema "github.com/nzhussup/konform/internal/schema"
)

// ReportEntry is a single resolved field in a LoadReport.
type ReportEntry struct {
	Path   string
	Value  string
	Source string
}

// LoadReport describes resolved values and their sources after loading.
type LoadReport struct {
	Entries []ReportEntry
}

// Print writes a tabular report to w.
func (r *LoadReport) Print(w io.Writer) {
	if r == nil || len(r.Entries) == 0 || w == nil {
		return
	}

	maxPathLen := 0
	maxValueLen := 0
	for _, entry := range r.Entries {
		if len(entry.Path) > maxPathLen {
			maxPathLen = len(entry.Path)
		}
		if len(entry.Value) > maxValueLen {
			maxValueLen = len(entry.Value)
		}
	}

	for _, entry := range r.Entries {
		_, _ = fmt.Fprintf(w, "%-*s = %-*s source=%s\n", maxPathLen, entry.Path, maxValueLen, entry.Value, entry.Source)
	}
}

func buildReport(sc *internalschema.Schema) *LoadReport {
	if sc == nil {
		return nil
	}

	pathAliases := internalschemaToLookupAliases(sc)
	entries := make([]ReportEntry, 0, len(sc.Fields))
	for _, field := range sc.Fields {
		if field.Type.Kind() == reflect.Struct {
			continue
		}

		var path string
		if field.KeyName != "" {
			path = field.KeyName
		} else {
			path = resolveLookupPathForReport(field, pathAliases)
		}

		source := field.Source
		if source == "" {
			source = "zero"
		}

		value := field.ValueAsString()
		if field.IsSecret {
			value = "***"
		}

		entries = append(entries, ReportEntry{
			Path:   path,
			Value:  value,
			Source: source,
		})
	}

	return &LoadReport{Entries: entries}
}

func internalschemaToLookupAliases(sc *internalschema.Schema) map[string]string {
	aliases := make(map[string]string, len(sc.Fields))
	for _, field := range sc.Fields {
		if field.KeyName == "" {
			continue
		}
		aliases[field.Path] = field.KeyName
	}
	return aliases
}

func resolveLookupPathForReport(field internalschema.Field, pathAliases map[string]string) string {
	if field.KeyName != "" {
		return field.KeyName
	}

	resolved := field.Path
	parts := strings.Split(field.Path, ".")
	for i := range parts {
		prefix := strings.Join(parts[:i+1], ".")
		alias, ok := pathAliases[prefix]
		if !ok {
			continue
		}

		suffix := strings.Join(parts[i+1:], ".")
		if suffix == "" {
			resolved = alias
		} else {
			resolved = alias + "." + suffix
		}
	}

	return resolved
}
