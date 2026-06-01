// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"strings"
	"testing"
)

func TestSwitchStmtAndExpr(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 2
    switch x:
        1:
            console.print_ln("one")
        _:
            console.print_ln("other")
    result: string = switch x:
        2 -> "two"
        _ -> "many"
    console.print_ln(result)
`)
	if out != "other\ntwo\n" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestRangeLoopVariants(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..3:
        console.print(i)
    console.print("|")
    loop i in 0..<3:
        console.print(i)
    console.print("|")
    loop i in 0..10, step 3:
        console.print(i)
`)
	if out != "0123|012|0369" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestLoopModifiersBreakContinueAndInfinite(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    n = 0
    loop i in 0..10, while n < 5, if i != 1:
        n = n + 1
        if i == 3:
            continue
        console.print(i)
        if i == 5:
            break
    console.print("|")
    count = 0
    loop true:
        count = count + 1
        if count == 3:
            break
    console.print(count)
`)
	if out != "0245|3" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestVarDeclWithZeroValue(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x: int
    y: float
    z: bool
    console.print_ln(x)
    console.print_ln(y)
    console.print_ln(z)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "false") {
		t.Fatalf("expected zero values, got %q", out)
	}
}

func TestSwitchExprNoMatch(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = switch 99:
        1 -> "one"
        2 -> "two"
    console.print_ln(x)
`)
	if !strings.Contains(out, "null") {
		t.Fatalf("expected null for no match, got %q", out)
	}
}

func TestSwitchStmtNoMatch(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 99
    switch x:
        1 -> console.print_ln("one")
        2 -> console.print_ln("two")
    console.print_ln("done")
`)
	if !strings.Contains(out, "done") {
		t.Fatalf("expected done, got %q", out)
	}
}

func TestLoopWithStep(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..10, step 2:
        console.print_ln(i)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "2") || !strings.Contains(out, "8") {
		t.Fatalf("expected 0, 2, 8, got %q", out)
	}
}

func TestLoopWithIfModifier(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..5, if i != 2:
        console.print_ln(i)
`)
	if strings.Contains(out, "2") {
		t.Fatalf("should not contain 2, got %q", out)
	}
}

func TestSwitchStmtWithDefault(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 99
    switch x:
        1 -> console.print_ln("one")
        _ -> console.print_ln("other")
`)
	if !strings.Contains(out, "other") {
		t.Fatalf("expected other, got %q", out)
	}
}

func TestSwitchExprWithDefault(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 99:
        1 -> "one"
        _ -> "other"
    console.print_ln(result)
`)
	if !strings.Contains(out, "other") {
		t.Fatalf("expected other, got %q", out)
	}
}

func TestElseIf(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 2
    if x == 1:
        console.print_ln("one")
    else if x == 2:
        console.print_ln("two")
    else:
        console.print_ln("other")
`)
	if !strings.Contains(out, "two") {
		t.Fatalf("expected two, got %q", out)
	}
}

func TestElse(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 3
    if x == 1:
        console.print_ln("one")
    else:
        console.print_ln("other")
`)
	if !strings.Contains(out, "other") {
		t.Fatalf("expected other, got %q", out)
	}
}

func TestLoopContinue(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..5:
        if i == 2:
            continue
        console.print_ln(i)
`)
	if strings.Contains(out, "2") {
		t.Fatalf("should not contain 2, got %q", out)
	}
}

func TestLoopBreak(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..10:
        if i == 3:
            break
        console.print_ln(i)
`)
	if strings.Contains(out, "3") {
		t.Fatalf("should not contain 3, got %q", out)
	}
}

func TestDictLoop(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": 1, "b": 2}
    loop k in m:
        console.print_ln(k)
`)
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("expected a and b, got %q", out)
	}
}

func TestSwitchStmtMultipleCases(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 2
    switch x:
        1 -> console.print_ln("one")
        2 -> console.print_ln("two")
        3 -> console.print_ln("three")
`)
	if !strings.Contains(out, "two") {
		t.Fatalf("expected two, got %q", out)
	}
}

func TestSwitchExprMultipleArms(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 3:
        1 -> "one"
        2 -> "two"
        3 -> "three"
    console.print_ln(result)
`)
	if !strings.Contains(out, "three") {
		t.Fatalf("expected three, got %q", out)
	}
}

func TestRangeLoopExclusive(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..<3:
        console.print_ln(i)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "2") || strings.Contains(out, "3") {
		t.Fatalf("expected 0, 1, 2 but not 3, got %q", out)
	}
}

func TestRangeLoopInclusive(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..3:
        console.print_ln(i)
`)
	if !strings.Contains(out, "3") {
		t.Fatalf("expected 3, got %q", out)
	}
}

func TestListLoop(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [10, 20, 30]
    loop x in xs:
        console.print_ln(x)
`)
	if !strings.Contains(out, "10") || !strings.Contains(out, "20") || !strings.Contains(out, "30") {
		t.Fatalf("expected list elements, got %q", out)
	}
}

func TestAssignDictKey(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": 1}
    m["b"] = 2
    console.print_ln(m["b"])
`)
	if !strings.Contains(out, "2") {
		t.Fatalf("expected 2, got %q", out)
	}
}

func TestAssignListElement(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [1, 2, 3]
    xs[0] = 99
    console.print_ln(xs[0])
`)
	if !strings.Contains(out, "99") {
		t.Fatalf("expected 99, got %q", out)
	}
}

// --- Additional coverage for remaining gaps ---

func TestSwitchStmtFallThrough(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 1
    switch x:
        1 -> console.print_ln("one")
        2 -> console.print_ln("two")
`)
	if !strings.Contains(out, "one") {
		t.Fatalf("expected one, got %q", out)
	}
}

func TestSwitchExprFirstMatch(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 1:
        1 -> "first"
        2 -> "second"
    console.print_ln(result)
`)
	if !strings.Contains(out, "first") {
		t.Fatalf("expected first, got %q", out)
	}
}

func TestAssignDictKeyNested(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": {"x": 1}}
    m["a"]["x"] = 99
    console.print_ln(m["a"]["x"])
`)
	if !strings.Contains(out, "99") {
		t.Fatalf("expected 99, got %q", out)
	}
}

func TestSwitchStmtNoMatchNoDefault(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 99
    switch x:
        1 -> console.print_ln("one")
    console.print_ln("done")
`)
	if !strings.Contains(out, "done") {
		t.Fatalf("expected done, got %q", out)
	}
}

func TestSwitchExprNoMatchNoDefault(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 99:
        1 -> "one"
    console.print_ln(result)
`)
	if !strings.Contains(out, "null") {
		t.Fatalf("expected null, got %q", out)
	}
}

func TestFieldAccessOnVariable(t *testing.T) {
	out := runSource(t, `package app.main

type Item:
    Name: string
    Value: int

func main():
    item = Item(Name: "test", Value: 42)
    n = item.Name
    v = item.Value
    console.print_ln(n)
    console.print_ln(v)
`)
	if !strings.Contains(out, "test") || !strings.Contains(out, "42") {
		t.Fatalf("expected test and 42, got %q", out)
	}
}

// --- Targeted coverage for specific branches ---

func TestLoopWithStepModifier(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..10, step 3:
        console.print_ln(i)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "3") || !strings.Contains(out, "9") {
		t.Fatalf("expected 0, 3, 6, 9, got %q", out)
	}
}

func TestLoopWithWhileModifier(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..10, while i < 5:
        console.print_ln(i)
`)
	if strings.Contains(out, "5") {
		t.Fatalf("should not contain 5, got %q", out)
	}
}

func TestRangeLoopWithDictValues(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": 1, "b": 2, "c": 3}
    loop v in m:
        console.print_ln(v)
`)
	if !strings.Contains(out, "1") || !strings.Contains(out, "2") || !strings.Contains(out, "3") {
		t.Fatalf("expected 1, 2, 3, got %q", out)
	}
}

func TestSwitchExprWithWildcard(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 99:
        1 -> "one"
        _ -> "default"
    console.print_ln(result)
`)
	if !strings.Contains(out, "default") {
		t.Fatalf("expected default, got %q", out)
	}
}

func TestSwitchStmtWithWildcard(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 99
    switch x:
        1 -> console.print_ln("one")
        _ -> console.print_ln("default")
`)
	if !strings.Contains(out, "default") {
		t.Fatalf("expected default, got %q", out)
	}
}

// --- Additional control flow tests ---

func TestDictIteration(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": 1, "b": 2}
    loop k in m:
        console.print_ln(k)
`)
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("expected a and b, got %q", out)
	}
}

func TestListIteration(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [10, 20, 30]
    loop x in xs:
        console.print_ln(x)
`)
	if !strings.Contains(out, "10") || !strings.Contains(out, "20") || !strings.Contains(out, "30") {
		t.Fatalf("expected 10, 20, 30, got %q", out)
	}
}

func TestSwitchStmtWithWildcardPattern(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 2
    switch x:
        1 -> console.print_ln("one")
        2 -> console.print_ln("two")
        _ -> console.print_ln("other")
`)
	if !strings.Contains(out, "two") {
		t.Fatalf("expected two, got %q", out)
	}
}

func TestSwitchExprWithWildcardPattern(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 3:
        1 -> "one"
        2 -> "two"
        _ -> "other"
    console.print_ln(result)
`)
	if !strings.Contains(out, "other") {
		t.Fatalf("expected other, got %q", out)
	}
}

func TestIfElseIfElse(t *testing.T) {
	out := runSource(t, `package app.main

func classify(x: int) -> string:
    if x < 0:
        return "negative"
    else if x == 0:
        return "zero"
    else:
        return "positive"

func main():
    console.print_ln(classify(-1))
    console.print_ln(classify(0))
    console.print_ln(classify(1))
`)
	if !strings.Contains(out, "negative") || !strings.Contains(out, "zero") || !strings.Contains(out, "positive") {
		t.Fatalf("expected negative, zero, positive, got %q", out)
	}
}

func TestLoopWithStepAndWhile(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..100, step 3, while i < 10:
        console.print_ln(i)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "3") || !strings.Contains(out, "9") || strings.Contains(out, "12") {
		t.Fatalf("expected 0, 3, 6, 9 but not 12, got %q", out)
	}
}

func TestLoopWithIfAndContinue(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..5, if i != 2:
        console.print_ln(i)
`)
	if strings.Contains(out, "2") {
		t.Fatalf("should not contain 2, got %q", out)
	}
}

func TestDictLoopEntry(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"name": "test", "value": "42"}
    loop entry in m:
        console.print_ln(entry)
`)
	if !strings.Contains(out, "name") || !strings.Contains(out, "test") {
		t.Fatalf("expected dict entries, got %q", out)
	}
}

