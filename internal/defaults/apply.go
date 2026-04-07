package defaults

import (
	"fmt"

	"github.com/nzhussup/conform/internal/decode"
	"github.com/nzhussup/conform/internal/errs"
	"github.com/nzhussup/conform/internal/schema"
)

func Apply(sc *schema.Schema) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	for _, f := range sc.Fields {
		if !f.HasDefaultValue {
			continue
		}

		if !schema.IsZeroValue(f.Value) {
			continue
		}

		if err := decode.SetFieldValue(f, f.DefaultValue); err != nil {
			return fmt.Errorf("%w: invalid default for %s: %w", errs.Decode, f.Path, err)
		}
	}

	return nil
}
