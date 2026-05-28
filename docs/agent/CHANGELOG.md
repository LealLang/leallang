# Changelog

All notable changes to the LealLang design specification will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

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
