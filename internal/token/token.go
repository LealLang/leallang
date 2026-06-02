// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

// Package token defines the token types for the LealLang lexer.
package token

// TokenKind represents the type of a lexical token.
type TokenKind uint8

const (
	// Special
	ILLEGAL TokenKind = iota
	EOF

	// Structural
	NEWLINE
	INDENT
	DEDENT

	// Literals
	INT_LIT
	FLOAT_LIT
	STRING_LIT
	INTERP_STRING_LIT
	CHAR_LIT

	// Identifier
	IDENT

	// Keywords
	PACKAGE
	IMPORT
	AS
	PUB
	FUNC
	RETURN
	IF
	ELSE
	SWITCH
	LOOP
	BREAK
	CONTINUE
	IN
	STEP
	WHILE
	PASS
	TYPE
	CONST
	REF
	AND
	OR
	NOT
	NULL
	TRUE
	FALSE
	ANY
	STRING_KW
	INT_KW
	FLOAT_KW
	BOOL_KW
	CHAR_KW
	ASYNC
	AWAIT
	UI
	ON

	// Operators
	PLUS
	MINUS
	STAR
	SLASH
	PERCENT
	EQ_EQ
	BANG_EQ
	LESS
	LESS_EQ
	GREATER
	GREATER_EQ
	EQ
	ARROW
	DOT_DOT
	DOT_DOT_LESS
	PIPE
	QUESTION
	AT
	DOT
	UNDERSCORE

	// Delimiters
	COMMA
	COLON
	LPAREN
	RPAREN
	LBRACKET
	RBRACKET
	LBRACE
	RBRACE
	LESS_ANGLE
	GREATER_ANGLE
)

var kindNames = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	NEWLINE: "NEWLINE",
	INDENT:  "INDENT",
	DEDENT:  "DEDENT",

	INT_LIT:           "INT_LIT",
	FLOAT_LIT:         "FLOAT_LIT",
	STRING_LIT:        "STRING_LIT",
	INTERP_STRING_LIT: "INTERP_STRING_LIT",
	CHAR_LIT:          "CHAR_LIT",

	IDENT: "IDENT",

	PACKAGE:   "PACKAGE",
	IMPORT:    "IMPORT",
	AS:        "AS",
	PUB:       "PUB",
	FUNC:      "FUNC",
	RETURN:    "RETURN",
	IF:        "IF",
	ELSE:      "ELSE",
	SWITCH:    "SWITCH",
	LOOP:      "LOOP",
	BREAK:     "BREAK",
	CONTINUE:  "CONTINUE",
	IN:        "IN",
	STEP:      "STEP",
	WHILE:     "WHILE",
	PASS:      "PASS",
	TYPE:      "TYPE",
	CONST:     "CONST",
	REF:       "REF",
	AND:       "AND",
	OR:        "OR",
	NOT:       "NOT",
	NULL:      "NULL",
	TRUE:      "TRUE",
	FALSE:     "FALSE",
	ANY:       "ANY",
	STRING_KW: "STRING_KW",
	INT_KW:    "INT_KW",
	FLOAT_KW:  "FLOAT_KW",
	BOOL_KW:   "BOOL_KW",
	CHAR_KW:   "CHAR_KW",
	ASYNC:     "ASYNC",
	AWAIT:     "AWAIT",
	UI:        "UI",
	ON:        "ON",

	PLUS:         "PLUS",
	MINUS:        "MINUS",
	STAR:         "STAR",
	SLASH:        "SLASH",
	PERCENT:      "PERCENT",
	EQ_EQ:        "EQ_EQ",
	BANG_EQ:      "BANG_EQ",
	LESS:         "LESS",
	LESS_EQ:      "LESS_EQ",
	GREATER:      "GREATER",
	GREATER_EQ:   "GREATER_EQ",
	EQ:           "EQ",
	ARROW:        "ARROW",
	DOT_DOT:      "DOT_DOT",
	DOT_DOT_LESS: "DOT_DOT_LESS",
	PIPE:         "PIPE",
	QUESTION:     "QUESTION",
	AT:           "AT",
	DOT:          "DOT",
	UNDERSCORE:   "UNDERSCORE",

	COMMA:         "COMMA",
	COLON:         "COLON",
	LPAREN:        "LPAREN",
	RPAREN:        "RPAREN",
	LBRACKET:      "LBRACKET",
	RBRACKET:      "RBRACKET",
	LBRACE:        "LBRACE",
	RBRACE:        "RBRACE",
	LESS_ANGLE:    "LESS_ANGLE",
	GREATER_ANGLE: "GREATER_ANGLE",
}

// String returns the human-readable name of the token kind.
func (k TokenKind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "UNKNOWN"
}

// keywords maps keyword strings to their TokenKind.
var keywords = map[string]TokenKind{
	"package":  PACKAGE,
	"import":   IMPORT,
	"as":       AS,
	"pub":      PUB,
	"func":     FUNC,
	"return":   RETURN,
	"if":       IF,
	"else":     ELSE,
	"switch":   SWITCH,
	"loop":     LOOP,
	"break":    BREAK,
	"continue": CONTINUE,
	"in":       IN,
	"step":     STEP,
	"while":    WHILE,
	"pass":     PASS,
	"type":     TYPE,
	"const":    CONST,
	"ref":      REF,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
	"null":     NULL,
	"true":     TRUE,
	"false":    FALSE,
	"any":      ANY,
	"string":   STRING_KW,
	"int":      INT_KW,
	"float":    FLOAT_KW,
	"bool":     BOOL_KW,
	"char":     CHAR_KW,
	"async":    ASYNC,
	"await":    AWAIT,
	"ui":       UI,
}

// LookupKeyword returns the TokenKind for a keyword, or IDENT if it is not a keyword.
func LookupKeyword(ident string) TokenKind {
	if kind, ok := keywords[ident]; ok {
		return kind
	}
	return IDENT
}

// Position represents a location in source code.
type Position struct {
	File   string
	Line   int // 1-based
	Col    int // 1-based byte column
	Offset int // byte offset from file start
}

// Token represents a lexical token.
type Token struct {
	Kind   TokenKind
	Lexeme string
	Pos    Position
}

// String returns a debug representation of the token.
func (t Token) String() string {
	if t.Lexeme != "" {
		return t.Kind.String() + " " + `"` + t.Lexeme + `"`
	}
	return t.Kind.String()
}
