package httpx

import (
	"strings"
	"testing"

	"github.com/Im-Doh/revu/backend/internal/decimal"
)

func ptr(s string) *string { return &s }

func TestValidatorMessageFormat(t *testing.T) {
	var v Validator
	v.NotBlank("title", ptr("   "), "Title must not be blank")
	v.Required("status", false, "Status must not be null")
	err := v.Err()
	if err == nil || err.Error() != "title: Title must not be blank, status: Status must not be null" {
		t.Fatalf("unexpected: %v", err)
	}
	if err.(*Error).Status != 400 {
		t.Fatal("expected 400")
	}
}

func TestEmail(t *testing.T) {
	good := []string{"a@b", "max.mustermann@example.de", "x+y@sub.domain.org"}
	bad := []string{"plain", "@example.com", "a@", "a b@c.de", "a@@b.de", "a@.de", ".a@b.de"}
	for _, e := range good {
		if !looksLikeEmail(e) {
			t.Errorf("%q should be valid", e)
		}
	}
	for _, e := range bad {
		if looksLikeEmail(e) {
			t.Errorf("%q should be invalid", e)
		}
	}
}

func TestLengthCountsCharacters(t *testing.T) {
	var v Validator
	s := strings.Repeat("ä", 255)
	v.MaxLen("name", &s, 255, "too long")
	if v.Err() != nil {
		t.Fatal("255 umlauts are 255 characters")
	}
}

func TestDecimalChecks(t *testing.T) {
	var v Validator
	small, big := decimal.MustParse("0.001"), decimal.MustParse("10000")
	v.DecimalMin("amount", &small, decimal.MustParse("0.01"), "min")
	v.DecimalFits("weightKg", &big, 6, 2, "max")
	if err := v.Err(); err == nil || err.Error() != "amount: min, weightKg: max" {
		t.Fatalf("unexpected %v", err)
	}
}
