package validate

import (
	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func MissingRequired(sc *schema.Schema) ([]schema.Field, error) {
	if sc == nil {
		return nil, errs.InvalidSchemaNil
	}

	var missing []schema.Field
	for _, f := range sc.Fields {
		if !f.Required {
			continue
		}
		if schema.IsZeroValue(f.Value) {
			missing = append(missing, f)
		}
	}

	return missing, nil
}
