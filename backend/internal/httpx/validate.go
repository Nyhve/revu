package httpx

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/Nyhve/revu/backend/internal/decimal"
)

// Validator collects field errors in declaration order and renders them as
// "field: message, field: message" (the format of the former Spring backend).
type Validator struct {
	errs []string
}

func (v *Validator) add(field, msg string) {
	v.errs = append(v.errs, field+": "+msg)
}

// Err returns a 400 error if any check failed, otherwise nil.
func (v *Validator) Err() error {
	if len(v.errs) == 0 {
		return nil
	}
	return BadRequest("%s", strings.Join(v.errs, ", "))
}

// NotBlank fails for missing, empty and whitespace-only strings.
func (v *Validator) NotBlank(field string, value *string, msg string) bool {
	if value == nil || strings.TrimSpace(*value) == "" {
		v.add(field, msg)
		return false
	}
	return true
}

// MaxLen fails if value has more than max characters (nil passes).
func (v *Validator) MaxLen(field string, value *string, max int, msg string) {
	if value != nil && utf8.RuneCountInString(*value) > max {
		v.add(field, msg)
	}
}

// MinLen fails if value has fewer than min characters (nil passes).
func (v *Validator) MinLen(field string, value *string, min int, msg string) {
	if value != nil && utf8.RuneCountInString(*value) < min {
		v.add(field, msg)
	}
}

// Email fails for strings that do not look like an e-mail address (nil/empty pass).
func (v *Validator) Email(field string, value *string, msg string) {
	if value == nil || *value == "" {
		return
	}
	if !looksLikeEmail(*value) {
		v.add(field, msg)
	}
}

// Required fails if the value is missing (nil).
func (v *Validator) Required(field string, present bool, msg string) bool {
	if !present {
		v.add(field, msg)
		return false
	}
	return true
}

// OneOf fails if value is set but not one of allowed.
func (v *Validator) OneOf(field string, value *string, allowed []string) {
	if value != nil && !slices.Contains(allowed, *value) {
		v.add(field, "must be one of "+strings.Join(allowed, ", "))
	}
}

// Min fails if value < min.
func (v *Validator) Min(field string, value, min int, msg string) {
	if value < min {
		v.add(field, msg)
	}
}

// Max fails if value > max.
func (v *Validator) Max(field string, value, max int, msg string) {
	if value > max {
		v.add(field, msg)
	}
}

// DecimalMin fails if value is set and below min (inclusive bound).
func (v *Validator) DecimalMin(field string, value *decimal.Decimal, min decimal.Decimal, msg string) {
	if value != nil && value.Cmp(min) < 0 {
		v.add(field, msg)
	}
}

// DecimalFits fails if value does not fit into NUMERIC(precision, scale).
func (v *Validator) DecimalFits(field string, value *decimal.Decimal, precision, scale int, msg string) {
	if value != nil && value.IntegerDigits() > precision-scale {
		v.add(field, msg)
	}
}

// looksLikeEmail mirrors the permissive Hibernate @Email check:
// a non-empty local part, a single '@' and a non-empty domain without spaces.
func looksLikeEmail(s string) bool {
	at := strings.LastIndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	if strings.ContainsAny(s, " \t\r\n") || strings.Count(s, "@") != 1 {
		return false
	}
	domain := s[at+1:]
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") || strings.Contains(domain, "..") {
		return false
	}
	local := s[:at]
	return !strings.HasPrefix(local, ".") && !strings.HasSuffix(local, ".")
}
