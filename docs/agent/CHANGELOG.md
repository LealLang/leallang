# Changelog

All notable changes to the LealLang design specification will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased] - 28-05-2026

### Added
- Updated `cmd/leal` CLI to parse `.ll` files using the AST parser by default
- Added `--tokens` flag to `cmd/leal` for raw token output
- Added `examples/full.ll` with types, constructors, switch expressions, loops, and interpolated strings

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
