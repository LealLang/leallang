// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package token

import "testing"

func TestLookupKeyword(t *testing.T) {
	tests := []struct {
		input string
		want  TokenKind
	}{
		{"package", PACKAGE},
		{"func", FUNC},
		{"return", RETURN},
		{"if", IF},
		{"else", ELSE},
		{"loop", LOOP},
		{"break", BREAK},
		{"continue", CONTINUE},
		{"type", TYPE},
		{"const", CONST},
		{"and", AND},
		{"or", OR},
		{"not", NOT},
		{"null", NULL},
		{"true", TRUE},
		{"false", FALSE},
		{"any", ANY},
		{"string", STRING_KW},
		{"int", INT_KW},
		{"float", FLOAT_KW},
		{"bool", BOOL_KW},
		{"char", CHAR_KW},
		{"my_variable", IDENT},
		{"count", IDENT},
		{"step_count", IDENT},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := LookupKeyword(tt.input); got != tt.want {
				t.Fatalf("LookupKeyword(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTokenKindString(t *testing.T) {
	tests := []struct {
		kind TokenKind
		want string
	}{
		{EOF, "EOF"},
		{IDENT, "IDENT"},
		{FUNC, "FUNC"},
		{PLUS, "PLUS"},
		{LPAREN, "LPAREN"},
		{LESS_ANGLE, "LESS_ANGLE"},
		{GREATER_ANGLE, "GREATER_ANGLE"},
		{TokenKind(255), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Fatalf("TokenKind.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	t.Run("with_lexeme", func(t *testing.T) {
		tok := Token{Kind: IDENT, Lexeme: "foo"}
		want := `IDENT "foo"`
		if got := tok.String(); got != want {
			t.Fatalf("Token.String() = %q, want %q", got, want)
		}
	})
	t.Run("without_lexeme", func(t *testing.T) {
		tok := Token{Kind: NEWLINE}
		want := "NEWLINE"
		if got := tok.String(); got != want {
			t.Fatalf("Token.String() = %q, want %q", got, want)
		}
	})
}
