// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

// Package lexer implements a streaming lexer for the LealLang programming language.
package lexer

import (
	"fmt"

	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/token"
)

const indentUnit = 4

// Lexer tokenizes LealLang source code into a stream of tokens.
type Lexer struct {
	src            string
	pos            int
	line           int
	col            int
	file           string
	diag           *diagnostics.Diagnostics
	indentStack    []int
	atLineStart    bool
	pendingDedents int
}

// New creates a new Lexer for the given source code.
func New(file, src string, diag *diagnostics.Diagnostics) *Lexer {
	return &Lexer{
		src:         src,
		pos:         0,
		line:        1,
		col:         1,
		file:        file,
		diag:        diag,
		indentStack: []int{0},
		atLineStart: true,
	}
}

// Tokenize returns all tokens from the source, including a trailing EOF.
func (l *Lexer) Tokenize() []token.Token {
	var tokens []token.Token
	for {
		t := l.NextToken()
		tokens = append(tokens, t)
		if t.Kind == token.EOF {
			break
		}
	}
	return tokens
}

// NextToken returns the next token from the source.
func (l *Lexer) NextToken() token.Token {
	// Emit any pending DEDENT tokens
	if l.pendingDedents > 0 {
		l.pendingDedents--
		return l.makeToken(token.DEDENT, "", l.pos)
	}

	// Handle indentation at the start of a logical line
	if l.atLineStart {
		return l.handleLineStart()
	}

	// Skip mid-line whitespace
	l.skipSpaces()

	if l.atEnd() {
		return l.emitEOF()
	}

	start := l.pos
	ch := l.advance()

	switch ch {
	case '\n':
		l.atLineStart = true
		return l.makeToken(token.NEWLINE, "", start)

	case '#':
		l.skipLineComment()
		// After a line comment, we're at the end of the line
		if !l.atEnd() && l.peek() == '\n' {
			l.advance()
			l.atLineStart = true
			return l.makeToken(token.NEWLINE, "", start)
		}
		// Comment at EOF
		return l.emitEOF()

	case '"':
		return l.scanString(start)

	case '\'':
		return l.scanChar(start)

	case '(':
		return l.makeToken(token.LPAREN, "(", start)
	case ')':
		return l.makeToken(token.RPAREN, ")", start)
	case '[':
		return l.makeToken(token.LBRACKET, "[", start)
	case ']':
		return l.makeToken(token.RBRACKET, "]", start)
	case '{':
		return l.makeToken(token.LBRACE, "{", start)
	case '}':
		return l.makeToken(token.RBRACE, "}", start)
	case ',':
		return l.makeToken(token.COMMA, ",", start)
	case ':':
		return l.makeToken(token.COLON, ":", start)
	case '?':
		return l.makeToken(token.QUESTION, "?", start)
	case '@':
		return l.makeToken(token.AT, "@", start)
	case '|':
		return l.makeToken(token.PIPE, "|", start)
	case '%':
		return l.makeToken(token.PERCENT, "%", start)
	case '+':
		return l.makeToken(token.PLUS, "+", start)
	case '*':
		return l.makeToken(token.STAR, "*", start)

	case '-':
		if l.peek() == '>' {
			l.advance()
			return l.makeToken(token.ARROW, "->", start)
		}
		return l.makeToken(token.MINUS, "-", start)

	case '=':
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(token.EQ_EQ, "==", start)
		}
		return l.makeToken(token.EQ, "=", start)

	case '!':
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(token.BANG_EQ, "!=", start)
		}
		l.diag.ReportError("E001", "unexpected character '!'",
			l.posAt(start), 0, "", l.currentLine())
		return l.makeToken(token.ILLEGAL, "!", start)

	case '<':
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(token.LESS_EQ, "<=", start)
		}
		return l.makeToken(token.LESS, "<", start)

	case '>':
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(token.GREATER_EQ, ">=", start)
		}
		return l.makeToken(token.GREATER, ">", start)

	case '.':
		if l.peek() == '.' {
			l.advance()
			if l.peek() == '<' {
				l.advance()
				return l.makeToken(token.DOT_DOT_LESS, "..<", start)
			}
			return l.makeToken(token.DOT_DOT, "..", start)
		}
		return l.makeToken(token.DOT, ".", start)

	case '/':
		if l.peek() == '*' {
			l.advance()
			l.scanBlockComment()
			// After block comment, skip whitespace and try again
			return l.NextToken()
		}
		return l.makeToken(token.SLASH, "/", start)

	case '$':
		if l.peek() == '"' {
			l.advance()
			return l.scanInterpString(start)
		}
		l.diag.ReportError("E001", "unexpected character '$'",
			l.posAt(start), 0, "", l.currentLine())
		return l.makeToken(token.ILLEGAL, "$", start)

	default:
		if ch == '_' && (l.atEnd() || !isAlphaNumeric(l.peek())) {
			return l.makeToken(token.UNDERSCORE, "_", start)
		}
		if isLetter(ch) || ch == '_' {
			return l.scanIdentifier(start)
		}
		if isDigit(ch) {
			return l.scanNumber(start)
		}
		l.diag.ReportError("E001", fmt.Sprintf("unexpected character '%c'", ch),
			l.posAt(start), 0, "", l.currentLine())
		return l.makeToken(token.ILLEGAL, string(ch), start)
	}
}

// handleLineStart processes indentation at the beginning of a logical line.
func (l *Lexer) handleLineStart() token.Token {
	start := l.pos
	for {
		l.atLineStart = false

		// Count leading spaces (do NOT skip them first — they are the indentation)
		spaces := 0
		for !l.atEnd() && l.peek() == ' ' {
			spaces++
			l.advance()
		}

		// Blank line or comment-only line: skip and don't emit indent tokens
		if l.atEnd() || l.peek() == '\n' {
			if !l.atEnd() {
				l.advance() // consume \n
			}
			if l.atEnd() {
				return l.emitEOF()
			}
			l.atLineStart = true
			continue
		}

		if l.peek() == '#' {
			l.skipLineComment()
			if !l.atEnd() && l.peek() == '\n' {
				l.advance()
			}
			if l.atEnd() {
				return l.emitEOF()
			}
			l.atLineStart = true
			continue
		}

		// Check for tab indentation
		if l.peek() == '\t' {
			l.diag.ReportError("E006", "tab used for indentation; spaces required",
				l.posAt(l.pos), 0, "use 4 spaces per indentation level", l.currentLine())
			l.advance()
			return l.makeToken(token.ILLEGAL, "\t", start)
		}

		// Check for non-multiple of 4
		if spaces%indentUnit != 0 {
			l.diag.ReportError("E007",
				fmt.Sprintf("indentation is %d spaces, not a multiple of %d", spaces, indentUnit),
				l.posAt(l.pos-spaces), 0, "use 4 spaces per indentation level", l.currentLine())
		}

		current := l.indentStack[len(l.indentStack)-1]

		if spaces > current {
			// Deeper indentation
			l.indentStack = append(l.indentStack, spaces)
			return l.makeToken(token.INDENT, "", start)
		}

		if spaces < current {
			// Shallower indentation — queue dedents
			count := 0
			for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > spaces {
				l.indentStack = l.indentStack[:len(l.indentStack)-1]
				count++
			}
			if l.indentStack[len(l.indentStack)-1] != spaces {
				l.diag.ReportError("E008", "indentation does not match any outer level",
					l.posAt(l.pos-spaces), 0, "", l.currentLine())
			}
			if count > 0 {
				l.pendingDedents = count - 1
				return l.makeToken(token.DEDENT, "", start)
			}
		}

		// Same indentation level — continue to normal token scanning
		return l.NextToken()
	}
}

// scanIdentifier reads an identifier or keyword token.
func (l *Lexer) scanIdentifier(start int) token.Token {
	for !l.atEnd() && isAlphaNumeric(l.peek()) {
		l.advance()
	}
	lexeme := l.src[start:l.pos]
	kind := token.LookupKeyword(lexeme)
	return l.makeToken(kind, lexeme, start)
}

// scanNumber reads an integer or float literal.
func (l *Lexer) scanNumber(start int) token.Token {
	for !l.atEnd() && isDigit(l.peek()) {
		l.advance()
	}

	// Check for float: digit followed by . digit
	if !l.atEnd() && l.peek() == '.' {
		next := l.peekAt(1)
		if isDigit(next) {
			l.advance() // consume '.'
			for !l.atEnd() && isDigit(l.peek()) {
				l.advance()
			}
			return l.makeToken(token.FLOAT_LIT, l.src[start:l.pos], start)
		}
	}

	return l.makeToken(token.INT_LIT, l.src[start:l.pos], start)
}

// scanString reads a string literal starting after the opening quote.
func (l *Lexer) scanString(start int) token.Token {
	for !l.atEnd() && l.peek() != '"' {
		if l.peek() == '\n' {
			l.diag.ReportError("E002", "unterminated string literal",
				l.posAt(start), 0, "add a closing \" before the end of the line", l.currentLine())
			return l.makeToken(token.STRING_LIT, l.src[start:l.pos], start)
		}
		if l.peek() == '\\' {
			l.advance() // consume backslash
			if !l.atEnd() {
				esc := l.peek()
				if esc != '"' && esc != '\\' && esc != 'n' && esc != 't' && esc != 'r' && esc != '0' && esc != '\'' {
					l.diag.ReportError("E005", fmt.Sprintf("invalid escape sequence '\\%c'", esc),
						l.posAt(l.pos-1), 0, "", l.currentLine())
				}
				l.advance()
			}
		} else {
			l.advance()
		}
	}

	if l.atEnd() {
		l.diag.ReportError("E002", "unterminated string literal",
			l.posAt(start), 0, "add a closing \" to end the string", l.currentLine())
		return l.makeToken(token.STRING_LIT, l.src[start:l.pos], start)
	}

	l.advance() // consume closing quote
	return l.makeToken(token.STRING_LIT, l.src[start:l.pos], start)
}

// scanInterpString reads an interpolated string starting after the $" prefix.
func (l *Lexer) scanInterpString(start int) token.Token {
	depth := 0
	for !l.atEnd() {
		ch := l.peek()
		if ch == '\n' {
			l.diag.ReportError("E002", "unterminated interpolated string literal",
				l.posAt(start), 0, "add a closing \" before the end of the line", l.currentLine())
			return l.makeToken(token.INTERP_STRING_LIT, l.src[start:l.pos], start)
		}
		if ch == '{' {
			depth++
			l.advance()
		} else if ch == '}' {
			depth--
			l.advance()
		} else if ch == '"' {
			l.advance()
			if depth == 0 {
				return l.makeToken(token.INTERP_STRING_LIT, l.src[start:l.pos], start)
			}
		} else if ch == '\\' {
			l.advance() // consume backslash
			if !l.atEnd() {
				l.advance() // consume escaped char
			}
		} else {
			l.advance()
		}
	}

	l.diag.ReportError("E002", "unterminated interpolated string literal",
		l.posAt(start), 0, "add a closing \" to end the string", l.currentLine())
	return l.makeToken(token.INTERP_STRING_LIT, l.src[start:l.pos], start)
}

// scanChar reads a char literal starting after the opening quote.
func (l *Lexer) scanChar(start int) token.Token {
	if l.atEnd() || l.peek() == '\n' {
		l.diag.ReportError("E009", "empty char literal",
			l.posAt(start), 0, "char literals must contain exactly one character", l.currentLine())
		return l.makeToken(token.CHAR_LIT, l.src[start:l.pos], start)
	}

	// Read the character or escape sequence
	if l.peek() == '\\' {
		l.advance() // backslash
		if !l.atEnd() {
			esc := l.peek()
			if esc != 'n' && esc != '\\' && esc != '\'' && esc != 't' && esc != 'r' && esc != '0' {
				l.diag.ReportError("E005", fmt.Sprintf("invalid escape sequence '\\%c'", esc),
					l.posAt(l.pos-1), 0, "", l.currentLine())
			}
			l.advance()
		}
	} else {
		l.advance()
	}

	// Expect closing quote
	if l.atEnd() || l.peek() != '\'' {
		if l.atEnd() || l.peek() == '\n' {
			l.diag.ReportError("E003", "unterminated char literal",
				l.posAt(start), 0, "add a closing ' to end the char literal", l.currentLine())
		} else {
			// Multiple characters
			for !l.atEnd() && l.peek() != '\'' && l.peek() != '\n' {
				l.advance()
			}
			if !l.atEnd() && l.peek() == '\'' {
				l.advance()
			}
			l.diag.ReportError("E010", "char literal contains more than one character",
				l.posAt(start), 0, "char literals must contain exactly one character", l.currentLine())
		}
		return l.makeToken(token.CHAR_LIT, l.src[start:l.pos], start)
	}

	l.advance() // consume closing quote
	return l.makeToken(token.CHAR_LIT, l.src[start:l.pos], start)
}

// scanBlockComment reads a block comment until */, reporting an error if unterminated.
// Supports nested block comments for robustness.
func (l *Lexer) scanBlockComment() {
	depth := 1
	for !l.atEnd() {
		if l.peek() == '/' && l.peekAt(1) == '*' {
			depth++
			l.advance()
			l.advance()
		} else if l.peek() == '*' && l.peekAt(1) == '/' {
			depth--
			l.advance()
			l.advance()
			if depth == 0 {
				return
			}
		} else {
			l.advance()
		}
	}
	l.diag.ReportError("E004", "unterminated multi-line comment",
		l.posAt(l.pos), 0, "add a closing */ to end the comment", l.currentLine())
}

// skipLineComment advances past a line comment (started by #) to end of line.
func (l *Lexer) skipLineComment() {
	for !l.atEnd() && l.peek() != '\n' {
		l.advance()
	}
}

// skipSpaces advances past space characters (not tabs, not newlines).
func (l *Lexer) skipSpaces() {
	for !l.atEnd() && l.peek() == ' ' {
		l.advance()
	}
}

// emitEOF handles end-of-file: emit remaining DEDENTs then EOF.
func (l *Lexer) emitEOF() token.Token {
	start := l.pos
	// Emit DEDENTs for all open levels (except base 0)
	if len(l.indentStack) > 1 {
		count := len(l.indentStack) - 1
		l.indentStack = l.indentStack[:1] // keep only base 0
		l.pendingDedents = count - 1
		return l.makeToken(token.DEDENT, "", start)
	}
	return l.makeToken(token.EOF, "", start)
}

// Character classification helpers

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlphaNumeric(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '_'
}

// Peek and navigation helpers

func (l *Lexer) peek() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peekAt(offset int) byte {
	pos := l.pos + offset
	if pos >= len(l.src) {
		return 0
	}
	return l.src[pos]
}

func (l *Lexer) atEnd() bool {
	return l.pos >= len(l.src)
}

func (l *Lexer) advance() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	ch := l.src[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) posAt(offset int) token.Position {
	// Compute line and column from the source using the offset
	line := 1
	for i := 0; i < offset && i < len(l.src); i++ {
		if l.src[i] == '\n' {
			line++
		}
	}
	// Find the start of this line
	lineStart := offset
	for lineStart > 0 && l.src[lineStart-1] != '\n' {
		lineStart--
	}
	col := offset - lineStart + 1 // 1-based column
	return token.Position{
		File:   l.file,
		Line:   line,
		Col:    col,
		Offset: offset,
	}
}

func (l *Lexer) currentLine() string {
	// Find the start of the current line
	start := l.pos
	for start > 0 && l.src[start-1] != '\n' {
		start--
	}
	end := l.pos
	for end < len(l.src) && l.src[end] != '\n' {
		end++
	}
	return l.src[start:end]
}

func (l *Lexer) makeToken(kind token.TokenKind, lexeme string, start int) token.Token {
	return token.Token{
		Kind:   kind,
		Lexeme: lexeme,
		Pos:    l.posAt(start),
	}
}
