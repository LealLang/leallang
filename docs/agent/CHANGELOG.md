# Changelog

All notable changes to the LealLang design specification will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased] - 28-05-2026

### Added
- Go compiler Phase 3 type checker implementation:
  - `internal/checker` package: semantic analysis with type inference, name resolution, and type validation
  - `types.go`: internal type system (PrimitiveType, NullType, NullableType, RecordType, FuncSignature, GenericType, TupleType, NamespaceType, ComponentType, EnumType) with assignability rules
  - `scope.go`: scope chain with symbol table (Define, Lookup, LookupLocal) supporting nested scopes
  - `builtins.go`: pre-registered built-in types (string, int, float, bool, char, any, Error), event types (ClickEvent, ResizeEvent, etc.), enum-like types (Color, Dock, Orientation, ToastType), 10 namespaces (console, file, window, msg, json, modal, toast, system, clipboard, screen), 8 constant groups, and 37 component types (Window, Button, Label, Panel, TextInput, etc.)
  - `checker.go`: two-pass analysis — Pass 1 collects top-level declarations, Pass 2 checks bodies; supports var/const declarations, function declarations, type declarations with constructors and methods
  - `check_expr.go` (inline): expression type inference for literals, identifiers, binary/unary operators, function calls (with positional + named args, ref param validation), field access, indexing, list/dict literals, ranges, interpolated strings, switch expressions, component refs
  - `check_stmt.go` (inline): statement checking for assignments (with implicit variable declarations), return statements, if/else (with bool condition validation), switch, loops (with range/List/Dict iteration, step/while/if modifiers)
  - Type checker diagnostic codes E021-E051 covering: undefined names, type mismatches, null safety, operator errors, call validation, record construction, immutability, ref params, visibility, return types, control flow, component validation
  - 55 unit tests covering all checker functionality
  - Default constructor generation for records without explicit constructors
- AST changes for type checker support:
  - Added `Pub bool` field to `ast.FieldDecl` (was parsed but discarded)
  - Changed `ast.FuncDecl.ReturnType` to `ReturnTypes []TypeExpr` for multi-return type support
  - Updated parser to pass field visibility through to AST
  - Updated parser to parse comma-separated return types
  - Updated AST pretty-printer for new ReturnTypes and FieldDecl.Pub

- Updated `cmd/leal` CLI to parse `.ll` files using the AST parser by default
- Added `--tokens` flag to `cmd/leal` for raw token output
- Added `examples/full.ll` with types, constructors, switch expressions, loops, and interpolated strings
- Added unit tests covering all compiler diagnostic codes E001-E020:
  - Lexer: E001 (unexpected `!`, `$`), E002 (interpolated string branches), E005 (invalid char escape), E008 (unmatched indentation), E009 (empty char exact code)
  - Parser: E012 (unterminated interpolation brace), E013 (const missing value), E014 (invalid assignment target), E015 (else without if), E016 (continue outside loop), E017 (duplicate while/if modifiers), E018 (empty switch), E019 (package not first), E020 (multiple packages)
- Added `internal/diagnostics/diagnostics_test.go` covering Severity.String, ReportError/ReportWarning, All/Errors/HasErrors, and Format output
- Added `internal/token/token_test.go` covering LookupKeyword, TokenKind.String, and Token.String
- Added GitHub Actions CI pipelines:
  - `build.yml`: build pipeline that runs first on every push and PR to main/release
  - `tests.yml`: unit test pipeline with each package (token, diagnostics, lexer, parser, AST) as a separate step, runs after build passes

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
