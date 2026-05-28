# Bugs to Fix

## Bug 1: `scanBlockComment` double-increments line counter

**File:** `internal/lexer/lexer.go`, lines 456-460

**Problem:** When the lexer encounters a newline inside a block comment, it manually increments `l.line` and resets `l.col` before calling `l.advance()`. But `advance()` (line 528) already handles newlines the same way. Every newline inside a block comment increments `l.line` twice.

**Code:**

```go
} else {
    if l.peek() == '\n' {
        l.line++
        l.col = 0 // will be incremented by advance
    }
    l.advance()
}
```

**Proof:** After a 4-line block comment (lines 1-4), a token `x` on line 5 reports as line 8 (off by 3, one per newline in the comment).

**Fix:** Remove the manual `l.line++` / `l.col = 0` and let `advance()` handle it:

```go
} else {
    l.advance()
}
```

---

## Bug 2: `makeToken` records end-of-token position, not start

**File:** `internal/lexer/lexer.go`, lines 578-589

**Problem:** `makeToken` uses `l.line`, `l.col`, and `l.pos` directly. By the time it's called, the lexer has already advanced past the token. Every token's `Pos` field points to the character _after_ the token, not its first character.

**Code:**

```go
func (l *Lexer) makeToken(kind token.TokenKind, lexeme string) token.Token {
    return token.Token{
        Kind:   kind,
        Lexeme: lexeme,
        Pos: token.Position{
            File:   l.file,
            Line:   l.line,
            Col:    l.col,
            Offset: l.pos,
        },
    }
}
```

**Proof:** `package` at the start of a file reports col=8 instead of col=1. `app` at offset 8 reports col=12 instead of col=9.

**Fix:** Capture the start position before scanning each token, then pass it to `makeToken`. The lexer already has a `posAt(offset)` helper that recomputes line/col from a byte offset. Update `makeToken` to accept a start offset:

```go
func (l *Lexer) makeToken(kind token.TokenKind, lexeme string, start int) token.Token {
    return token.Token{
        Kind:   kind,
        Lexeme: lexeme,
        Pos:    l.posAt(start),
    }
}
```

Then in `NextToken()` and each `scan*` method, save `start := l.pos` before advancing and pass it to `makeToken`. The diagnostics system already uses `posAt(start)` correctly, so only the token positions are affected.
