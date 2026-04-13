package validate

import (
	"fmt"
	"sort"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
	"github.com/nzhussup/konform/internal/validate/types"
)

func Validate(sc *schema.Schema, registry map[string]types.ValidationFunc) ([]types.ValidationResult, error) {
	if sc == nil {
		return nil, errs.InvalidSchemaNil
	}

	var results []types.ValidationResult
	for _, f := range sc.Fields {
		if len(f.Validations) == 0 {
			continue
		}

		ruleNames := make([]string, 0, len(f.Validations))
		for ruleName := range f.Validations {
			ruleNames = append(ruleNames, ruleName)
		}
		sort.Strings(ruleNames)

		for _, ruleName := range ruleNames {
			validator, ok := registry[ruleName]
			if !ok {
				return nil, fmt.Errorf("%w: unsupported validate rule %q for field %q", errs.InvalidSchema, ruleName, f.Path)
			}
			validator(f, &results)
		}
	}

	return results, nil
}
