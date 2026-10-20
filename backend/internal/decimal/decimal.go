// Package decimal is a small exact decimal type for money and weights.
//
// It replaces Java's BigDecimal: values keep their scale, are stored as
// PostgreSQL NUMERIC and are serialised as plain JSON numbers (e.g. 12.50),
// just like Jackson did.
package decimal

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// Decimal represents coef * 10^exp.
type Decimal struct {
	coef *big.Int
	exp  int32
}

var Zero = Decimal{coef: big.NewInt(0)}

// Parse reads a plain or exponent notation decimal such as "12.5", "-3", "1e2".
func Parse(s string) (Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Decimal{}, fmt.Errorf("invalid decimal %q", s)
	}
	mantissa, expPart := s, int64(0)
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		mantissa = s[:i]
		var e big.Int
		if _, ok := e.SetString(s[i+1:], 10); !ok || !e.IsInt64() {
			return Decimal{}, fmt.Errorf("invalid decimal %q", s)
		}
		expPart = e.Int64()
	}
	digits := mantissa
	scale := int64(0)
	if i := strings.IndexByte(mantissa, '.'); i >= 0 {
		scale = int64(len(mantissa) - i - 1)
		digits = mantissa[:i] + mantissa[i+1:]
	}
	if digits == "" || digits == "-" || digits == "+" {
		return Decimal{}, fmt.Errorf("invalid decimal %q", s)
	}
	coef, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return Decimal{}, fmt.Errorf("invalid decimal %q", s)
	}
	exp := expPart - scale
	if exp > 1<<20 || exp < -(1<<20) {
		return Decimal{}, fmt.Errorf("decimal %q out of range", s)
	}
	return Decimal{coef: coef, exp: int32(exp)}, nil
}

func MustParse(s string) Decimal {
	d, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return d
}

func (d Decimal) c() *big.Int {
	if d.coef == nil {
		return new(big.Int)
	}
	return d.coef
}

func (d Decimal) rat() *big.Rat {
	r := new(big.Rat).SetInt(d.c())
	if d.exp == 0 {
		return r
	}
	pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(abs(d.exp))), nil)
	if d.exp > 0 {
		return r.Mul(r, new(big.Rat).SetInt(pow))
	}
	return r.Quo(r, new(big.Rat).SetInt(pow))
}

// Cmp compares d and o numerically (-1, 0, +1).
func (d Decimal) Cmp(o Decimal) int { return d.rat().Cmp(o.rat()) }

// Sub returns d - o, keeping the larger scale (like BigDecimal.subtract).
func (d Decimal) Sub(o Decimal) Decimal {
	a, b := align(d, o)
	return Decimal{coef: new(big.Int).Sub(a.c(), b.c()), exp: a.exp}
}

// Add returns d + o, keeping the larger scale.
func (d Decimal) Add(o Decimal) Decimal {
	a, b := align(d, o)
	return Decimal{coef: new(big.Int).Add(a.c(), b.c()), exp: a.exp}
}

// Scale returns the number of digits after the decimal point.
func (d Decimal) Scale() int32 { return -d.exp }

// IntegerDigits returns the number of digits before the decimal point.
func (d Decimal) IntegerDigits() int {
	s := new(big.Int).Abs(d.c()).String()
	n := len(s) + int(d.exp)
	if n < 0 {
		return 0
	}
	return n
}

func align(a, b Decimal) (Decimal, Decimal) {
	switch {
	case a.exp == b.exp:
		return a, b
	case a.exp > b.exp:
		return rescale(a, b.exp), b
	default:
		return a, rescale(b, a.exp)
	}
}

// rescale lowers the exponent of d to exp (exp must be <= d.exp).
func rescale(d Decimal, exp int32) Decimal {
	pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d.exp-exp)), nil)
	return Decimal{coef: new(big.Int).Mul(d.c(), pow), exp: exp}
}

// String renders plain notation, e.g. "12.50", "0", "-3.1".
func (d Decimal) String() string {
	coef := d.c()
	if d.exp >= 0 {
		if d.exp == 0 || coef.Sign() == 0 {
			if coef.Sign() == 0 {
				return "0"
			}
			return coef.String()
		}
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d.exp)), nil)
		return new(big.Int).Mul(coef, pow).String()
	}
	neg := coef.Sign() < 0
	digits := new(big.Int).Abs(coef).String()
	scale := int(-d.exp)
	if len(digits) <= scale {
		digits = strings.Repeat("0", scale-len(digits)+1) + digits
	}
	out := digits[:len(digits)-scale] + "." + digits[len(digits)-scale:]
	if neg {
		out = "-" + out
	}
	return out
}

func (d Decimal) MarshalJSON() ([]byte, error) { return []byte(d.String()), nil }

// UnmarshalJSON accepts JSON numbers and numeric strings (Jackson does both).
func (d *Decimal) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if bytes.Equal(b, []byte("null")) {
		return nil
	}
	s := string(bytes.Trim(b, `"`))
	v, err := Parse(s)
	if err != nil {
		return err
	}
	*d = v
	return nil
}

// Numeric converts to the pgx representation.
func (d Decimal) Numeric() pgtype.Numeric {
	return pgtype.Numeric{Int: new(big.Int).Set(d.c()), Exp: d.exp, Valid: true}
}

// FromNumeric converts from the pgx representation (NULL becomes zero).
func FromNumeric(n pgtype.Numeric) Decimal {
	if !n.Valid || n.Int == nil {
		return Zero
	}
	return Decimal{coef: new(big.Int).Set(n.Int), exp: n.Exp}
}

func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
