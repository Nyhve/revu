package decimal

import (
	"encoding/json"
	"testing"
)

func TestParseAndString(t *testing.T) {
	cases := map[string]string{
		"12.50": "12.50", "0": "0", "0.01": "0.01", "-3.1": "-3.1",
		"1e2": "100", "2500.5": "2500.5", "0.000": "0.000", "7": "7", "1.5E-1": "0.15",
	}
	for in, want := range cases {
		d, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		if got := d.String(); got != want {
			t.Errorf("Parse(%q).String() = %q, want %q", in, got, want)
		}
	}
	for _, bad := range []string{"", "abc", "1.2.3", "-", ".", "1e"} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) should fail", bad)
		}
	}
}

func TestArithmeticKeepsScale(t *testing.T) {
	a, b := MustParse("2500.50"), MustParse("120.25")
	if got := a.Sub(b).String(); got != "2380.25" {
		t.Fatalf("Sub = %s", got)
	}
	if got := Zero.Sub(MustParse("1.5")).String(); got != "-1.5" {
		t.Fatalf("Zero.Sub = %s", got)
	}
	if got := MustParse("1").Add(MustParse("0.25")).String(); got != "1.25" {
		t.Fatalf("Add = %s", got)
	}
}

func TestCmpAndDigits(t *testing.T) {
	if MustParse("0.01").Cmp(MustParse("0.010")) != 0 {
		t.Fatal("0.01 should equal 0.010")
	}
	if MustParse("0.009").Cmp(MustParse("0.01")) >= 0 {
		t.Fatal("0.009 < 0.01")
	}
	if n := MustParse("9999.99").IntegerDigits(); n != 4 {
		t.Fatalf("IntegerDigits = %d", n)
	}
	if n := MustParse("0.5").IntegerDigits(); n != 1 && n != 0 {
		t.Fatalf("IntegerDigits(0.5) = %d", n)
	}
}

func TestJSON(t *testing.T) {
	var v struct {
		A Decimal  `json:"a"`
		B Decimal  `json:"b"`
		C *Decimal `json:"c"`
	}
	if err := json.Unmarshal([]byte(`{"a": 12.5, "b": "80", "c": null}`), &v); err != nil {
		t.Fatal(err)
	}
	if v.A.String() != "12.5" || v.B.String() != "80" || v.C != nil {
		t.Fatalf("unexpected %v %v %v", v.A, v.B, v.C)
	}
	out, _ := json.Marshal(map[string]Decimal{"x": MustParse("12.50"), "z": Zero})
	if string(out) != `{"x":12.50,"z":0}` {
		t.Fatalf("marshal = %s", out)
	}
}

func TestNumericRoundTrip(t *testing.T) {
	d := MustParse("123.45")
	if got := FromNumeric(d.Numeric()).String(); got != "123.45" {
		t.Fatalf("round trip = %s", got)
	}
}
