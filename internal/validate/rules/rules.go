package rules

import (
	"github.com/nzhussup/konform/internal/validate/types"
	"github.com/nzhussup/konform/internal/validate/validators"
)

var Registry = map[string]types.ValidationFunc{
	validators.RequiredRuleName: validators.Required,
	validators.MinRuleName:      validators.Min,
	validators.MaxRuleName:      validators.Max,
	validators.LenRuleName:      validators.Len,
	validators.MinLenRuleName:   validators.MinLen,
	validators.MaxLenRuleName:   validators.MaxLen,
	validators.RegexRuleName:    validators.Regex,
	validators.OneOfRuleName:    validators.OneOf,
	validators.NonZeroRuleName:  validators.NonZero,
	validators.URLRuleName:      validators.URL,
	validators.EmailRuleName:    validators.Email,
}

func IsSupported(name string) bool {
	_, ok := Registry[name]
	return ok
}
