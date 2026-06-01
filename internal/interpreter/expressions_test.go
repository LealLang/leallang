// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"strings"
	"testing"
)

func TestUnaryMinusInt(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = -42
    console.print_ln(x)
`)
	if !strings.Contains(out, "-42") {
		t.Fatalf("expected -42, got %q", out)
	}
}

func TestUnaryMinusFloat(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = -3.14
    console.print_ln(x)
`)
	if !strings.Contains(out, "-3.14") {
		t.Fatalf("expected -3.14, got %q", out)
	}
}

func TestUnaryNot(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    a = not false
    b = not true
    console.print_ln(a)
    console.print_ln(b)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestUnaryMinusRuntimeError(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    x = -"hello"
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for unary minus on string")
	}
}

func TestNotRuntimeError(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    x = not 42
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for not on int")
	}
}

func TestFloatLiteral(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 3.14
    console.print_ln(x)
`)
	if !strings.Contains(out, "3.14") {
		t.Fatalf("expected 3.14, got %q", out)
	}
}

func TestCharLiteral(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 'A'
    console.print_ln(x)
`)
	if !strings.Contains(out, "A") {
		t.Fatalf("expected A, got %q", out)
	}
}

func TestBoolLiterals(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(true)
    console.print_ln(false)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestNullLiteral(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x: string? = null
    console.print_ln(x)
`)
	if !strings.Contains(out, "null") {
		t.Fatalf("expected null, got %q", out)
	}
}

func TestIntSubtraction(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(10 - 3)
`)
	if !strings.Contains(out, "7") {
		t.Fatalf("expected 7, got %q", out)
	}
}

func TestIntMultiplication(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(4 * 5)
`)
	if !strings.Contains(out, "20") {
		t.Fatalf("expected 20, got %q", out)
	}
}

func TestIntModulo(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(10 % 3)
`)
	if !strings.Contains(out, "1") {
		t.Fatalf("expected 1, got %q", out)
	}
}

func TestFloatArithmeticOps(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 + 2.5)
    console.print_ln(5.0 - 1.5)
    console.print_ln(2.0 * 3.5)
    console.print_ln(10.0 / 4.0)
`)
	if !strings.Contains(out, "4") || !strings.Contains(out, "3.5") || !strings.Contains(out, "7") || !strings.Contains(out, "2.5") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestFloatModulo(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(10.5 % 3.0)
`)
	if !strings.Contains(out, "1.5") {
		t.Fatalf("expected 1.5, got %q", out)
	}
}

func TestMixedIntFloatArithmetic(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 + 2.5)
`)
	if !strings.Contains(out, "3.5") {
		t.Fatalf("expected 3.5, got %q", out)
	}
}

func TestNumericComparisonAllOps(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 < 2)
    console.print_ln(2 > 1)
    console.print_ln(1 <= 1)
    console.print_ln(1 >= 1)
    console.print_ln(1.5 < 2.5)
    console.print_ln(2.5 > 1.5)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestLogicalAndShortCircuit(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(true and true)
    console.print_ln(true and false)
    console.print_ln(false and true)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestLogicalOrShortCircuit(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(false or true)
    console.print_ln(true or false)
    console.print_ln(false or false)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestEqualityOperators(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 == 1)
    console.print_ln(1 != 2)
    console.print_ln("a" == "a")
    console.print_ln("a" != "b")
    console.print_ln(true == true)
    console.print_ln(null == null)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestComparisonIntEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 == 1)
    console.print_ln(1 == 2)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonFloatEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 == 1.5)
    console.print_ln(1.5 == 2.5)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonIntLess(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 < 2)
    console.print_ln(2 < 1)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonIntGreater(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(2 > 1)
    console.print_ln(1 > 2)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonIntLessEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 <= 1)
    console.print_ln(1 <= 2)
    console.print_ln(2 <= 1)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonIntGreaterEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 >= 1)
    console.print_ln(2 >= 1)
    console.print_ln(1 >= 2)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonFloatLess(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 < 2.5)
    console.print_ln(2.5 < 1.5)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonFloatGreater(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(2.5 > 1.5)
    console.print_ln(1.5 > 2.5)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonFloatLessEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 <= 1.5)
    console.print_ln(1.5 <= 2.5)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestComparisonFloatGreaterEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 >= 1.5)
    console.print_ln(2.5 >= 1.5)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestUnaryMinusOnVariable(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 42
    y = -x
    console.print_ln(y)
`)
	if !strings.Contains(out, "-42") {
		t.Fatalf("expected -42, got %q", out)
	}
}

func TestUnaryNotOnVariable(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = true
    y = not x
    console.print_ln(y)
`)
	if !strings.Contains(out, "false") {
		t.Fatalf("expected false, got %q", out)
	}
}

func TestLogicalAndBothFalse(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(false and false)
`)
	if !strings.Contains(out, "false") {
		t.Fatalf("expected false, got %q", out)
	}
}

func TestLogicalOrBothTrue(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(true or true)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestIntArithmeticAllOps(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(10 + 3)
    console.print_ln(10 - 3)
    console.print_ln(10 * 3)
    console.print_ln(10 / 3)
    console.print_ln(10 % 3)
`)
	if !strings.Contains(out, "13") || !strings.Contains(out, "7") || !strings.Contains(out, "30") || !strings.Contains(out, "3") || !strings.Contains(out, "1") {
		t.Fatalf("expected arithmetic results, got %q", out)
	}
}

func TestFloatArithmeticAllOps(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 + 2.5)
    console.print_ln(5.0 - 1.5)
    console.print_ln(2.0 * 3.0)
    console.print_ln(10.0 / 4.0)
`)
	if !strings.Contains(out, "4") || !strings.Contains(out, "3.5") || !strings.Contains(out, "6") || !strings.Contains(out, "2.5") {
		t.Fatalf("expected float arithmetic results, got %q", out)
	}
}

