package guard

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Python's json.dump(obj, indent=2) and the number semantics beneath it, for
// a port whose stdout has to equal a Python script's byte for byte
// (design.md byte-identical-json). A value in a pyObj tree is one of nil
// (None), int, float64, string, pyObj or []pyObj; keeping Go's int and
// float64 apart is what keeps Python's int/float distinction — `3` against
// `3.0` — in the printed document.

// pyObj is a Python dict: its keys print in insertion order.
type pyObj []pyKV

type pyKV struct {
	k string
	v any
}

// pyFloat is Python's repr() of a float: the shortest digits that read back
// as f, in fixed notation from 1e-4 up to (not including) 1e16 and for zero,
// with ".0" added to an integral value; otherwise exponent notation with a
// signed, at least two-digit exponent (`1e-05`, `1e+16`), which is also Go's
// 'e' form.
func pyFloat(f float64) string {
	switch {
	case math.IsNaN(f):
		return "nan"
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	}
	if a := math.Abs(f); a != 0 && (a < 1e-4 || a >= 1e16) {
		return strconv.FormatFloat(f, 'e', -1, 64)
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

// pyRound is Python's round(x, n) on a float: x's exact binary value rounded
// to n decimals, half to even, then read back — what both CPython (via its
// dtoa) and strconv's correctly rounded 'f' format do. int(pyRound(x, 0)) is
// round(x) with ndigits omitted. -0.0 keeps its sign; NaN and ±Inf pass
// through.
func pyRound(x float64, n int) float64 {
	f, _ := strconv.ParseFloat(strconv.FormatFloat(x, 'f', n, 64), 64)
	return f
}

// pyDump is json.dumps(v, indent=2).
func pyDump(v any) string {
	var b strings.Builder
	pyWrite(&b, v, "")
	return b.String()
}

func pyWrite(b *strings.Builder, v any, indent string) {
	switch v := v.(type) {
	case nil:
		b.WriteString("null")
	case int:
		b.WriteString(strconv.Itoa(v))
	case float64:
		switch {
		case math.IsNaN(v):
			b.WriteString("NaN")
		case math.IsInf(v, 1):
			b.WriteString("Infinity")
		case math.IsInf(v, -1):
			b.WriteString("-Infinity")
		default:
			b.WriteString(pyFloat(v))
		}
	case string:
		// Every string this package dumps is a fixed key or a #rrggbb colour,
		// so nothing needs json's escaping.
		b.WriteString(`"` + v + `"`)
	case pyObj:
		if len(v) == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteString("{")
		for i, kv := range v {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString("\n" + indent + `  "` + kv.k + `": `)
			pyWrite(b, kv.v, indent+"  ")
		}
		b.WriteString("\n" + indent + "}")
	case []pyObj:
		if len(v) == 0 {
			b.WriteString("[]")
			return
		}
		b.WriteString("[")
		for i, o := range v {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString("\n" + indent + "  ")
			pyWrite(b, o, indent+"  ")
		}
		b.WriteString("\n" + indent + "]")
	default:
		panic(fmt.Sprintf("pyDump: %T is not a JSON value", v))
	}
}
