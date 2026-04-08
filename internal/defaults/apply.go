package defaults

import (
	"fmt"

	"github.com/nzhussup/konform/internal/decode"
	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func Apply(sc *schema.Schema) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	for i := range sc.Fields {
		field := sc.Fields[i]
		if !field.HasDefaultValue() {
			continue
		}

		if !schema.IsZeroValue(field.Value) {
			continue
		}

		if err := decode.SetFieldValue(field, field.DefaultValue); err != nil {
			ctx := fmt.Sprintf("invalid default for %s", field.Path)
			return errs.WrapDecode(errs.Decode, ctx, err)
		}
		sc.Fields[i].Source = "default"
	}

	return nil
}
