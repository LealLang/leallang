// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package lexer

import (
	"testing"

	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/token"
)

// collectTokens tokenizes the source and returns the token kinds (excluding EOF).
func collectTokens(t *testing.T, src string) []token.TokenKind {
	t.Helper()
	diag := diagnostics.New()
	l := New("test.ll", src, diag)
	tokens := l.Tokenize()
	var kinds []token.TokenKind
	for _, tok := range tokens {
		if tok.Kind != token.EOF {
			kinds = append(kinds, tok.Kind)
		}
	}
	return kinds
}

// collectTokensWithDiag returns tokens and diagnostics.
func collectTokensWithDiag(t *testing.T, src string) ([]token.Token, *diagnostics.Diagnostics) {
	t.Helper()
	diag := diagnostics.New()
	l := New("test.ll", src, diag)
	tokens := l.Tokenize()
	return tokens, diag
}

// expectTokens tokenizes src and asserts the token kinds match expected.
func expectTokens(t *testing.T, src string, expected []token.TokenKind) {
	t.Helper()
	actual := collectTokens(t, src)
	if len(actual) != len(expected) {
		t.Errorf("token count mismatch: got %d, want %d\ngot:  %v\nwant: %v",
			len(actual), len(expected), actual, expected)
		return
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Errorf("token[%d]: got %v, want %v\nfull got:  %v\nfull want: %v",
				i, actual[i], expected[i], actual, expected)
			return
		}
	}
}

// expectTokenLiterals tokenizes src and asserts lexemes match for non-EOF tokens.
func expectTokenLiterals(t *testing.T, src string, expected []struct {
	kind   token.TokenKind
	lexeme string
}) {
	t.Helper()
	diag := diagnostics.New()
	l := New("test.ll", src, diag)
	tokens := l.Tokenize()
	// Remove trailing EOF
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == token.EOF {
		tokens = tokens[:len(tokens)-1]
	}
	if len(tokens) != len(expected) {
		t.Errorf("token count mismatch: got %d, want %d", len(tokens), len(expected))
		for i, tok := range tokens {
			t.Logf("  [%d] %v", i, tok)
		}
		return
	}
	for i, exp := range expected {
		if tokens[i].Kind != exp.kind {
			t.Errorf("token[%d] kind: got %v, want %v", i, tokens[i].Kind, exp.kind)
		}
		if tokens[i].Lexeme != exp.lexeme {
			t.Errorf("token[%d] lexeme: got %q, want %q", i, tokens[i].Lexeme, exp.lexeme)
		}
	}
}

// --- Tests ---

func TestKeywords(t *testing.T) {
	tests := []struct {
		src  string
		kind token.TokenKind
	}{
		{"package", token.PACKAGE},
		{"import", token.IMPORT},
		{"as", token.AS},
		{"pub", token.PUB},
		{"func", token.FUNC},
		{"return", token.RETURN},
		{"if", token.IF},
		{"else", token.ELSE},
		{"switch", token.SWITCH},
		{"loop", token.LOOP},
		{"break", token.BREAK},
		{"continue", token.CONTINUE},
		{"in", token.IN},
		{"step", token.STEP},
		{"while", token.WHILE},
		{"pass", token.PASS},
		{"type", token.TYPE},
		{"const", token.CONST},
		{"ref", token.REF},
		{"and", token.AND},
		{"or", token.OR},
		{"not", token.NOT},
		{"null", token.NULL},
		{"true", token.TRUE},
		{"false", token.FALSE},
		{"any", token.ANY},
		{"string", token.STRING_KW},
		{"int", token.INT_KW},
		{"float", token.FLOAT_KW},
		{"bool", token.BOOL_KW},
		{"char", token.CHAR_KW},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectTokens(t, tt.src, []token.TokenKind{tt.kind})
		})
	}
}

func TestIdentifiers(t *testing.T) {
	tests := []struct {
		src  string
		kind token.TokenKind
	}{
		{"user_name", token.IDENT},
		{"x", token.IDENT},
		{"_test", token.IDENT},
		{"load_config", token.IDENT},
		{"main_window", token.IDENT},
		{"func_name", token.IDENT}, // not confused with FUNC keyword
		{"step_count", token.IDENT}, // not confused with STEP keyword
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectTokens(t, tt.src, []token.TokenKind{tt.kind})
		})
	}
}

func TestUnderscore(t *testing.T) {
	expectTokens(t, "_", []token.TokenKind{token.UNDERSCORE})
}

func TestIntegerLiterals(t *testing.T) {
	tests := []string{"42", "0", "1920", "100"}
	for _, src := range tests {
		t.Run(src, func(t *testing.T) {
			expectTokens(t, src, []token.TokenKind{token.INT_LIT})
		})
	}
}

func TestFloatLiterals(t *testing.T) {
	tests := []string{"9.99", "0.5", "1.0", "100.25"}
	for _, src := range tests {
		t.Run(src, func(t *testing.T) {
			expectTokens(t, src, []token.TokenKind{token.FLOAT_LIT})
		})
	}
}

func TestDotAfterInteger(t *testing.T) {
	// 1..3 should be INT_LIT DOT_DOT INT_LIT (not a float)
	expectTokens(t, "1..3", []token.TokenKind{
		token.INT_LIT, token.DOT_DOT, token.INT_LIT,
	})
}

func TestStringLiterals(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		expectTokens(t, `"hello"`, []token.TokenKind{token.STRING_LIT})
	})
	t.Run("empty", func(t *testing.T) {
		expectTokens(t, `""`, []token.TokenKind{token.STRING_LIT})
	})
	t.Run("with_escape_quote", func(t *testing.T) {
		expectTokens(t, `"he said \"hi\""`, []token.TokenKind{token.STRING_LIT})
	})
	t.Run("with_escape_backslash", func(t *testing.T) {
		expectTokens(t, `"path\\to"`, []token.TokenKind{token.STRING_LIT})
	})
	t.Run("with_newline_escape", func(t *testing.T) {
		expectTokens(t, `"line1\nline2"`, []token.TokenKind{token.STRING_LIT})
	})
}

func TestInterpolatedStrings(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		expectTokens(t, `$"hello"`, []token.TokenKind{token.INTERP_STRING_LIT})
	})
	t.Run("with_expression", func(t *testing.T) {
		expectTokens(t, `$"Hello, {name}"`, []token.TokenKind{token.INTERP_STRING_LIT})
	})
	t.Run("nested_braces", func(t *testing.T) {
		expectTokens(t, `$"result: {func()}"`, []token.TokenKind{token.INTERP_STRING_LIT})
	})
}

func TestCharLiterals(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		expectTokens(t, `'a'`, []token.TokenKind{token.CHAR_LIT})
	})
	t.Run("digit", func(t *testing.T) {
		expectTokens(t, `'0'`, []token.TokenKind{token.CHAR_LIT})
	})
	t.Run("escape_n", func(t *testing.T) {
		expectTokens(t, `'\n'`, []token.TokenKind{token.CHAR_LIT})
	})
	t.Run("escape_backslash", func(t *testing.T) {
		expectTokens(t, `'\\'`, []token.TokenKind{token.CHAR_LIT})
	})
	t.Run("escape_quote", func(t *testing.T) {
		expectTokens(t, `'\''`, []token.TokenKind{token.CHAR_LIT})
	})
}

func TestOperators(t *testing.T) {
	tests := []struct {
		src  string
		kind token.TokenKind
	}{
		{"+", token.PLUS},
		{"-", token.MINUS},
		{"*", token.STAR},
		{"/", token.SLASH},
		{"%", token.PERCENT},
		{"==", token.EQ_EQ},
		{"!=", token.BANG_EQ},
		{"<", token.LESS},
		{"<=", token.LESS_EQ},
		{">", token.GREATER},
		{">=", token.GREATER_EQ},
		{"=", token.EQ},
		{"->", token.ARROW},
		{"..", token.DOT_DOT},
		{"..<", token.DOT_DOT_LESS},
		{"|", token.PIPE},
		{"?", token.QUESTION},
		{"@", token.AT},
		{".", token.DOT},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectTokens(t, tt.src, []token.TokenKind{tt.kind})
		})
	}
}

func TestDelimiters(t *testing.T) {
	tests := []struct {
		src  string
		kind token.TokenKind
	}{
		{",", token.COMMA},
		{":", token.COLON},
		{"(", token.LPAREN},
		{")", token.RPAREN},
		{"[", token.LBRACKET},
		{"]", token.RBRACKET},
		{"{", token.LBRACE},
		{"}", token.RBRACE},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectTokens(t, tt.src, []token.TokenKind{tt.kind})
		})
	}
}

func TestOperatorCombinations(t *testing.T) {
	// Test that operators are not confused with each other
	tests := []struct {
		src     string
		kinds   []token.TokenKind
	}{
		{"===", []token.TokenKind{token.EQ_EQ, token.EQ}},
		{"!==", []token.TokenKind{token.BANG_EQ, token.EQ}},
		{"<<", []token.TokenKind{token.LESS, token.LESS}},
		{">>", []token.TokenKind{token.GREATER, token.GREATER}},
		{"->>", []token.TokenKind{token.ARROW, token.GREATER}},
		{"-+", []token.TokenKind{token.MINUS, token.PLUS}},
		{"*/", []token.TokenKind{token.STAR, token.SLASH}},
		{"..<=", []token.TokenKind{token.DOT_DOT_LESS, token.EQ}},
		{"...", []token.TokenKind{token.DOT_DOT, token.DOT}},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectTokens(t, tt.src, tt.kinds)
		})
	}
}

func TestSingleLineComments(t *testing.T) {
	t.Run("comment_alone", func(t *testing.T) {
		expectTokens(t, "# this is a comment", nil)
	})
	t.Run("comment_after_code", func(t *testing.T) {
		expectTokens(t, "x # comment\ny", []token.TokenKind{
			token.IDENT, token.NEWLINE, token.IDENT,
		})
	})
	t.Run("comment_on_own_line", func(t *testing.T) {
		// Comment-only line is skipped entirely (no NEWLINE emitted)
		expectTokens(t, "# comment\nx", []token.TokenKind{
			token.IDENT,
		})
	})
}

func TestMultiLineComments(t *testing.T) {
	t.Run("single_line_block", func(t *testing.T) {
		expectTokens(t, "x /* comment */ y", []token.TokenKind{
			token.IDENT, token.IDENT,
		})
	})
	t.Run("multi_line_block", func(t *testing.T) {
		expectTokens(t, "x /* line1\nline2 */ y", []token.TokenKind{
			token.IDENT, token.IDENT,
		})
	})
}

func TestBasicIndentation(t *testing.T) {
	src := "func foo():\n    pass\n"
	expectTokens(t, src, []token.TokenKind{
		token.FUNC, token.IDENT, token.LPAREN, token.RPAREN, token.COLON,
		token.NEWLINE,
		token.INDENT, token.PASS, token.NEWLINE,
		token.DEDENT,
	})
}

func TestNestedIndentation(t *testing.T) {
	src := "if x:\n    if y:\n        pass\n"
	expectTokens(t, src, []token.TokenKind{
		token.IF, token.IDENT, token.COLON,
		token.NEWLINE,
		token.INDENT,
		token.IF, token.IDENT, token.COLON,
		token.NEWLINE,
		token.INDENT,
		token.PASS, token.NEWLINE,
		token.DEDENT,
		token.DEDENT,
	})
}

func TestDedentAtEOF(t *testing.T) {
	src := "if x:\n    pass\n"
	expectTokens(t, src, []token.TokenKind{
		token.IF, token.IDENT, token.COLON,
		token.NEWLINE,
		token.INDENT, token.PASS, token.NEWLINE,
		token.DEDENT,
	})
}

func TestMultipleDedentsAtOnce(t *testing.T) {
	src := "if x:\n    if y:\n        pass\ny = 1\n"
	expectTokens(t, src, []token.TokenKind{
		token.IF, token.IDENT, token.COLON,
		token.NEWLINE,
		token.INDENT,
		token.IF, token.IDENT, token.COLON,
		token.NEWLINE,
		token.INDENT,
		token.PASS, token.NEWLINE,
		token.DEDENT,
		token.DEDENT,
		token.IDENT, token.EQ, token.INT_LIT, token.NEWLINE,
	})
}

func TestBlankLineBetweenBlocks(t *testing.T) {
	src := "if x:\n    pass\n\ny = 1\n"
	expectTokens(t, src, []token.TokenKind{
		token.IF, token.IDENT, token.COLON,
		token.NEWLINE,
		token.INDENT, token.PASS, token.NEWLINE,
		token.DEDENT,
		token.IDENT, token.EQ, token.INT_LIT, token.NEWLINE,
	})
}

func TestCommentOnlyLineBetweenBlocks(t *testing.T) {
	src := "if x:\n    # comment\n    pass\n"
	expectTokens(t, src, []token.TokenKind{
		token.IF, token.IDENT, token.COLON,
		token.NEWLINE,
		token.INDENT,
		token.PASS, token.NEWLINE,
		token.DEDENT,
	})
}

func TestErrorUnterminatedString(t *testing.T) {
	_, diag := collectTokensWithDiag(t, `"hello`)
	if !diag.HasErrors() {
		t.Fatal("expected errors for unterminated string")
	}
	errs := diag.Errors()
	if errs[0].Code != "E002" {
		t.Errorf("expected E002, got %s", errs[0].Code)
	}
}

func TestErrorTabIndentation(t *testing.T) {
	_, diag := collectTokensWithDiag(t, "func foo():\n\tpass\n")
	if !diag.HasErrors() {
		t.Fatal("expected errors for tab indentation")
	}
	errs := diag.Errors()
	if errs[0].Code != "E006" {
		t.Errorf("expected E006, got %s", errs[0].Code)
	}
}

func TestErrorBadEscape(t *testing.T) {
	_, diag := collectTokensWithDiag(t, `"hello\qworld"`)
	if !diag.HasErrors() {
		t.Fatal("expected errors for bad escape")
	}
	errs := diag.Errors()
	if errs[0].Code != "E005" {
		t.Errorf("expected E005, got %s", errs[0].Code)
	}
}

func TestErrorUnexpectedCharacter(t *testing.T) {
	_, diag := collectTokensWithDiag(t, "x ~ y")
	if !diag.HasErrors() {
		t.Fatal("expected errors for unexpected character")
	}
	errs := diag.Errors()
	if errs[0].Code != "E001" {
		t.Errorf("expected E001, got %s", errs[0].Code)
	}
}

func TestErrorUnterminatedChar(t *testing.T) {
	_, diag := collectTokensWithDiag(t, "'a")
	if !diag.HasErrors() {
		t.Fatal("expected errors for unterminated char")
	}
	errs := diag.Errors()
	if errs[0].Code != "E003" {
		t.Errorf("expected E003, got %s", errs[0].Code)
	}
}

func TestErrorEmptyChar(t *testing.T) {
	_, diag := collectTokensWithDiag(t, "''")
	if !diag.HasErrors() {
		t.Fatal("expected errors for empty char literal")
	}
}

func TestErrorMultiCharLiteral(t *testing.T) {
	_, diag := collectTokensWithDiag(t, "'ab'")
	if !diag.HasErrors() {
		t.Fatal("expected errors for multi-char literal")
	}
	errs := diag.Errors()
	found := false
	for _, e := range errs {
		if e.Code == "E010" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected E010 in errors, got %v", errs)
	}
}

func TestEndToEnd(t *testing.T) {
	src := `package app.main

func greet(name: string):
    console.print_ln($"Hello, {name}")
`
	expectTokens(t, src, []token.TokenKind{
		token.PACKAGE, token.IDENT, token.DOT, token.IDENT, // package app.main
		token.NEWLINE,
		// blank line does not emit NEWLINE
		token.FUNC, token.IDENT, token.LPAREN, // func greet(
		token.IDENT, token.COLON, token.STRING_KW, // name: string
		token.RPAREN, token.COLON, // ):
		token.NEWLINE,
		token.INDENT,
		token.IDENT, token.DOT, token.IDENT, // console.print_ln
		token.LPAREN,
		token.INTERP_STRING_LIT, // $"Hello, {name}"
		token.RPAREN,
		token.NEWLINE,
		token.DEDENT,
	})
}

func TestEmptySource(t *testing.T) {
	expectTokens(t, "", nil)
}

func TestOnlyNewlines(t *testing.T) {
	// Blank lines do not emit NEWLINE per spec
	expectTokens(t, "\n\n", nil)
}

func TestMixedExpression(t *testing.T) {
	src := `count > 0 and err == null`
	expectTokens(t, src, []token.TokenKind{
		token.IDENT, token.GREATER, token.INT_LIT,
		token.AND, token.IDENT, token.EQ_EQ, token.NULL,
	})
}

func TestCollectionIndexing(t *testing.T) {
	src := `numbers[0]`
	expectTokens(t, src, []token.TokenKind{
		token.IDENT, token.LBRACKET, token.INT_LIT, token.RBRACKET,
	})
}

func TestRangeOperators(t *testing.T) {
	t.Run("inclusive", func(t *testing.T) {
		expectTokens(t, "0..3", []token.TokenKind{
			token.INT_LIT, token.DOT_DOT, token.INT_LIT,
		})
	})
	t.Run("exclusive", func(t *testing.T) {
		expectTokens(t, "0..<3", []token.TokenKind{
			token.INT_LIT, token.DOT_DOT_LESS, token.INT_LIT,
		})
	})
}

func TestArrow(t *testing.T) {
	expectTokens(t, "->", []token.TokenKind{token.ARROW})
}

func TestTokenLexemes(t *testing.T) {
	expectTokenLiterals(t, `x 42 "hi"`, []struct {
		kind   token.TokenKind
		lexeme string
	}{
		{token.IDENT, "x"},
		{token.INT_LIT, "42"},
		{token.STRING_LIT, `"hi"`},
	})
}

func TestSwitchWithArrow(t *testing.T) {
	src := `switch color_name:
    "blue" -> colors.blue
    _ -> colors.white`
	expectTokens(t, src, []token.TokenKind{
		token.SWITCH, token.IDENT, token.COLON,
		token.NEWLINE,
		token.INDENT,
		token.STRING_LIT, token.ARROW, token.IDENT, token.DOT, token.IDENT,
		token.NEWLINE,
		token.UNDERSCORE, token.ARROW, token.IDENT, token.DOT, token.IDENT,
		token.DEDENT,
	})
}

func TestNestedBlockComments(t *testing.T) {
	src := `x /* outer /* inner */ still_comment */ y`
	expectTokens(t, src, []token.TokenKind{
		token.IDENT, token.IDENT,
	})
}

func TestUnterminatedBlockComment(t *testing.T) {
	_, diag := collectTokensWithDiag(t, `x /* never closed`)
	if !diag.HasErrors() {
		t.Fatal("expected errors for unterminated block comment")
	}
	errs := diag.Errors()
	if errs[0].Code != "E004" {
		t.Errorf("expected E004, got %s", errs[0].Code)
	}
}

func TestIndentationNotMultipleOf4(t *testing.T) {
	_, diag := collectTokensWithDiag(t, "func foo():\n  pass\n")
	if !diag.HasErrors() {
		t.Fatal("expected errors for non-multiple-of-4 indentation")
	}
	errs := diag.Errors()
	found := false
	for _, e := range errs {
		if e.Code == "E007" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected E007 in errors, got %v", errs)
	}
}
