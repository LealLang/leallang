# Changelog

All notable changes to the LealLang design specification will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased] - 28-05-2026

### Added
- Updated `cmd/leal` CLI to parse `.ll` files using the AST parser by default
- Added `--tokens` flag to `cmd/leal` for raw token output
- Added `examples/full.ll` with types, constructors, switch expressions, loops, and interpolated strings
- Added unit tests covering all compiler diagnostic codes E001-E020:
  - Lexer: E001 (unexpected `!`, `$`), E002 (interpolated string branches), E005 (invalid char escape), E008 (unmatched indentation), E009 (empty char exact code)
  - Parser: E012 (unterminated interpolation brace), E013 (const missing value), E014 (invalid assignment target), E015 (else without if), E016 (continue outside loop), E017 (duplicate while/if modifiers), E018 (empty switch), E019 (package not first), E020 (multiple packages)
- Added `internal/diagnostics/diagnostics_test.go` covering Severity.String, ReportError/ReportWarning, All/Errors/HasErrors, and Format output
- Added `internal/token/token_test.go` covering LookupKeyword, TokenKind.String, and Token.String

## [0] - 27-05-2006

### Added

- Go compiler Phase 2 parser implementation:
  - `internal/ast` package: AST node definitions for programs, declarations, statements, expressions, type annotations, and a debug pretty-printer
  - `internal/parser` package: recursive descent/Pratt parser from lexer token streams to AST programs
  - Parser diagnostics E011-E020 for argument ordering, interpolation braces, expected tokens, assignment targets, control-flow misuse, loop modifiers, switches, and package placement
  - Parser tests covering declarations, expressions, statements, switches, loops, interpolation, recovery, pretty-printing, and an end-to-end snippet
  - Updated `cmd/leal` CLI to parse `.ll` files and print AST by default, with `--tokens` flag for raw token output

- Go compiler Phase 1 implementation:
  - `internal/token` package: TokenKind enum (60+ kinds), Token struct, Position struct, keyword lookup
  - `internal/diagnostics` package: Diagnostic collector with Error/Warning severity, source-excerpt formatting with carets, error codes E001-E010
  - `internal/lexer` package: Streaming lexer with `NextToken()` API supporting all LealLang lexical forms — identifiers, keywords, literals (int, float, string, char, interpolated string), operators, indentation-based scoping (INDENT/DEDENT), comments (single-line and block with nesting), error recovery
  - `go.mod` module definition (`github.com/LealLang/leallang`)

- Complete LealLang language design reference covering:
  - Core language: design goals, lexical structure, expressions, packages, types
  - Functions: declarations, arguments, multiple returns, error handling, control flow
  - Components: windows, widgets, containers, complex components, events
  - Built-in namespaces: console, file, window, msg, json, modal, toast
  - Built-in constants: colors, dock, orientation, toast_type
  - Plugin system with Lua, HTML, and CSS
  - Application structure and full example

### Fixed

- Lexer: `scanBlockComment` double-incremented line counter on newlines inside block comments
- Lexer: `makeToken` recorded end-of-token position instead of start position

### Changed

- Reorganized docs into structured subdirectories (core/, functions/, components/, builtins/, plugins/, examples/)
- Tightened syntax reference across all spec sections
