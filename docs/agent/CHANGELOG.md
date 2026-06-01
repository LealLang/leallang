# Changelog

All notable changes to the LealLang design specification will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.0.1] - 01-06-2026

### Fixed

- Deduplicated error code `E013` which was used as a catch-all for 17+ distinct errors across parser and checker:
  - `internal/parser/parser.go`: Assigned unique codes E080–E095 for parser-specific errors (async/ui mutual exclusion, ui func outside Window, expected tokens, component block validation, interpolation, etc.)
  - `internal/checker/checker.go`: Assigned E075 (component reference outside ui func) and E076 (component declaration outside ui function)
  - `internal/parser/parser_test.go`: Updated 5 test expectations to use new codes
  - `internal/checker/checker_test.go`: Updated 2 test expectations to use new codes
- Parser infinite loop when bare `func` appears inside component blocks:
  - `internal/parser/parser.go`: `synchronize()` now advances past stop-tokens when `previous` is NEWLINE, preventing the error-recovery loop from getting stuck at the same token position
  - `internal/parser/parser.go`: `parseComponentDecl()` now handles bare `func` declarations inside Window blocks (alongside `ui func`), storing them in `comp.Funcs` with `UI: false`
  - `internal/interpreter/interpreter.go`: `registerComponentFuncs()` uses `fn.UI` instead of hardcoded `true` so bare funcs are registered as non-UI
- Checker: `@ComponentRef` syntax is now only allowed inside `ui func` bodies:
  - `internal/checker/checker.go`: `checkComponentRef()` rejects `@Component[id]` when `inUIFunc` is false (except `@Window[id]` which is allowed as a value)
  - `internal/checker/checker.go`: `checkFieldExpr()` and `checkCallExpr()` handle cross-window calls (`@Window[id].func()`) specially, allowing them even outside `ui func` bodies
  - `examples/test.ll`: Updated to use `ui func handle_resize()` with `@Window[main].w`/`.h`, added `func handle_hover()` as bare func example, added `window.open(@Window[main])` call in `main()`
- Checker: `window.open()` and `window.close()` no longer require `ref` parameter:
  - `internal/checker/builtins.go`: Changed parameter from `{Name: "ref", Type: AnyType, Ref: true}` to `{Name: "window_ref", Type: AnyType}` to match spec syntax `window.open(@Window[id])`

### Implemented

- First-class UI support as a complete language feature:
  - `internal/ast/ast.go`: `ComponentDecl`, `ComponentProp`, `EventBinding` AST nodes with source positions and interface markers
  - `internal/token/token.go`: `ON` keyword token for event binding syntax
  - `internal/parser/parser.go`: `parseUIBlock()` and `parseComponentDecl()` methods for parsing `ComponentType[id]:` blocks with props, events, and nested children inside `ui func` bodies
  - `internal/checker/checker.go`: `checkComponentDecl()` with full validation — component type lookup (35 types), duplicate ID detection (E070), property validation (E072), type checking (E022), event validation (E073), handler signature validation (E074)
  - `internal/uiir/uiir.go`: Stable UI intermediate representation (`Op`, `OpKind`, `Log`) independent of AST/checker/desktop framework
  - `internal/uiir/fakebackend.go`: Test backend that records mount, prop-set, and event-bind operations for headless testing
  - `internal/interpreter/interpreter.go`: `evalComponentDecl()` evaluates component trees, records UI operations to backend log, supports `@Component[id].prop = value` assignments via `OpPropSet`
  - `internal/interpreter/value.go`: `valueToUI()` helper for LealLang-to-Go value conversion
  - 20 new tests: 7 parser, 8 checker, 5 interpreter integration

### Changed

- `internal/interpreter/ui_test.go`: Updated all 4 UI integration tests to use new top-level `Window[...]` pattern instead of `ui func view():` wrapper. UI operations (mount, prop-set, event-bind) now happen during program initialization rather than when `main()` runs. `main()` uses `pass` or calls handlers directly.
- Enforced `ui func` must be inside `Window[...]` blocks:
  - `internal/ast/ast.go`: Added `Funcs []*FuncDecl` field to `ComponentDecl` for storing `ui func` declarations inside Window blocks
  - `internal/parser/parser.go`: Added `windowDepth` tracking, `ui func` parsing inside component blocks, rejection of `ui func` outside Window blocks and at top level, top-level `ComponentDecl` parsing support, `peekNext()` helper
  - `internal/checker/checker.go`: Added `collectComponentDecls()` for collecting `ui func` from component trees, `checkComponentDeclTopLevel()` for top-level Window validation, `checkComponentDeclInWindow()` for components inside Window's tree with duplicate ID checking
  - `internal/interpreter/interpreter.go`: Added `registerComponentFuncs()` for registering `ui func` from component trees, top-level `ComponentDecl` evaluation in program initialization
  - `internal/ast/print.go`: Updated `ComponentDecl` pretty-print to include nested `Funcs`
  - `internal/parser/parser_test.go`: Updated all component tests to use `Window[...]` at top level, added `TestUIFuncInsideWindowBlock`, `TestUIFuncOutsideWindowBlock`, `TestUIFuncInsideNonWindowComponent`
  - `internal/checker/checker_test.go`: Updated all component tests to use `Window[...]` at top level
- Updated spec docs to match block syntax:
  - `docs/leallang/components/index.md`: Component declarations use `ComponentType[id]:` with `prop = value` and `on event = handler`
  - `docs/leallang/components/widgets.md`: All widget examples use block syntax
  - `docs/leallang/components/containers.md`: All container examples use block syntax
  - `docs/leallang/components/complex.md`: All complex component examples use block syntax
  - `docs/leallang/components/events.md`: Event binding uses `on event = handler` syntax
  - `docs/leallang/core/lexical.md`: Updated component syntax examples
  - `docs/leallang/core/expressions.md`: Updated component declaration examples, removed `@Window[id].ComponentType[id]` form
  - `docs/leallang/core/packages.md`: Updated Window/component visibility examples
  - `docs/leallang/core/types.md`: Updated Label example to block syntax
  - `docs/leallang/builtins/constants.md`: Updated Panel examples to block syntax
  - `docs/leallang/plugins/index.md`: Updated ColorPicker example to block syntax
  - `docs/leallang/examples/full-example.md`: Complete rewrite with `Window[...]` at top level and `ui func` inside
  - `docs/leallang/examples/application-structure.md`: Updated Window/Button examples to block syntax
  - `docs/leallang/functions/index.md`: Updated `ui func` example with Window block pattern
- `examples/test.ll`: Restructured to use `Window[main]:` at top level with `ui func handle_save()` inside
- `internal/checker/checker.go`: Added `windowMethods` map for tracking `ui func` methods per Window instance, `@Window[id].func()` cross-window call resolution via `ComponentRefExpr` field lookup
- `internal/interpreter/interpreter.go`: Added `ui func` method lookup in `evalFieldExpr` for `ComponentRefVal` — resolves `@Window[id].func()` calls by looking up registered functions in the environment
- Split `internal/interpreter/interpreter_test.go` (3,220 lines) into 8 focused test files:
  - `test_helpers_test.go` (102 lines) — shared test infrastructure (runSource, safeWriter, etc.)
  - `async_test.go` (506 lines) — 26 tests for async/await, TaskVal, and cloneForTask
  - `value_test.go` (307 lines) — 11 tests for value types, Signal, JSON conversion
  - `expressions_test.go` (444 lines) — 35 tests for arithmetic, comparisons, logical ops
  - `control_flow_test.go` (608 lines) — 41 tests for switch, loops, if/else
  - `errors_and_builtins_test.go` (458 lines) — 36 tests for runtime errors and builtins
  - `records_and_functions_test.go` (700 lines) — 45 tests for records, functions, strings
  - `declarations_test.go` (170 lines) — 12 tests for const declarations and stub namespaces

## [0.0.1] - 31-05-2026

### Added

- 15 interpreter tests for async/await and Task behaviors:
  - 4 TaskVal unit tests: resolve/await, error resolve, type/string, await blocks until resolve
  - 6 cloneForTask tests: primitives, null/enum, range, tuple deep-copy, nested structures, all non-sendable types rejected
  - 5 async behavioral tests: void function, nested async calls, globals isolation, record-with-list deep copy, multiple awaits on same task
- 7 checker tests for async/await:
  - E013 async+ui mutual exclusion, E061 UI component ref rejection in async
  - console/file allowed in async (non-UI-affine), async calling async allowed
  - void async return type handling, await in non-async function on Task
- ~80 interpreter coverage tests to reach 85% statement coverage:
  - Value Type()/String() methods for all 20+ value types
  - valuesEqual for all type combinations (int, float, string, bool, char, null, enum, list, dict, record, tuple, component ref, range)
  - dictKey for all valid types (string, int, bool, char, enum) and invalid types
  - valueToGo/goToValue round-trip conversions
  - decodeJSONValue for all JSON types (null, bool, string, int, float, array, object)
  - jsonNumberToValue for nested arrays, objects, and default fallback
  - Signal.Error() for all signal kinds (return, break, continue, unknown)
  - evalUnary: -int, -float, not bool, error paths
  - evalLiteral: float, char, bool, null literals
  - numericBinary: subtraction, multiplication, modulo, float ops, mixed int/float
  - compareValues: all comparison operators on int and float
  - evalField: record field access, namespace member, const group member
  - evalIndex: string index, dict index, index out of bounds
  - evalAssign: identifier, index, record field, new variable, const reassignment
  - assignIndex: list and dict index assignment
  - evalSwitchStmt/evalSwitchExpr: match, no match, wildcard, multiple cases
  - evalIf: else-if, else branches
  - evalLoop: step, while, if modifiers, break, continue, dict/list iteration
  - builtins: file.write_text, file.exists, file.read_text error, json.parse success/error, json.stringify, system.env, system.args, console.print, stub namespace calls (window, msg, modal, toast)
  - Constructor with explicit body, default field values, partial args
  - Record methods with arguments, mutation, return values
  - Ref parameters, nested function calls, recursive functions
  - Call depth exceeded (E107), const declaration types (int, float, string, bool)
  - Interpolated strings with expressions, string concatenation
  - cloneForTask with non-sendable elements in list, dict, tuple containers
  - discardWriter nil path

### Fixed

- Async task runtime isolation: each `async func` now runs on a child `*Interpreter` with its own `globals`, `callDepth`, `exitCode`, `diag`, and `builtins` instead of sharing the parent's mutable state.
- Async argument cloning: all arguments to `async func` are deep-copied via `cloneForTask` before crossing the task boundary. Lists, dicts, records, and tuples are recursively cloned. Non-sendable values (functions, builtins, namespaces, etc.) are rejected with a clear error.
- Async closure isolation: `FuncVal` entries in the task environment are cloned with closures re-bound to the task's global snapshot, preventing async tasks from reaching into the caller's local scope chain.
- `TaskVal` reshaped to store both `result` and `err` (Null on success, Error record on task failure). `resolve(result, err)` and `await() (Value, Value)` signatures.
- `await` now returns `(T, Error?)` tuple at runtime, aligning with the language spec. Checker's `checkAwaitExpr` updated to return `TupleType{T, Error?}` for `Task<T>` and `Error?` for void tasks.
- Thread-safe I/O: shared `stdout`/`stderr` writers are wrapped with a `syncWriter` mutex so concurrent async tasks don't race on output.
- Checker: `checkAsyncSafety` now also runs on `FieldExpr` callees in `checkCallExpr`, fixing E062 detection for `window.open()`, `msg.error()`, etc. inside async functions.

### Added

- `cloneForTask(v Value) (Value, error)` — deep-copies sendable values and rejects non-sendable ones at task boundaries.
- `(*Interpreter).NewChild(taskEnv *Env) *Interpreter` — creates an isolated child interpreter for async task execution.
- `(*Env).SnapshotGlobals() *Env` — creates a flat environment containing only top-level bindings.
- `syncWriter` type for thread-safe concurrent writes to shared I/O writers.
- 8 interpreter tests for async: returns before completion, concurrent tasks, runtime error propagation, list/dict/record deep-copy, non-sendable rejection, callDepth isolation.
- 5 checker tests for async: E063 ref param rejection, E062 UI-affine namespace rejection, E064 await non-task rejection, async call returns Task, await returns payload+error.

## [0.0.1] - 31-05-2026

### Added

- Integrated threading/concurrency design into language spec:
  - `docs/leallang/functions/index.md`: `async func`, `ui func`, `await` sections with safety rules
  - `docs/leallang/core/types.md`: `Task<T>` type and sendable value rules for task boundaries
  - `docs/leallang/core/expressions.md`: `await` expressions and cross-window `@Window[id].ui_func()` calls
  - `docs/leallang/components/index.md`: window actors, window-owned `ui func` declarations, cross-window communication patterns
  - `docs/leallang/components/events.md`: async event handlers that may suspend at `await`
  - `docs/leallang/builtins/window.md`: non-blocking `window.open`, close cancellation behavior, concurrency error codes (`window_closed`, `task_cancelled`, `task_failed`)

### Implemented

- Threading/concurrency compiler support across all layers:
  - `internal/token/token.go`: `ASYNC`, `AWAIT`, `UI` keyword tokens
  - `internal/ast/ast.go`: `AwaitExpr` node, `Async`/`UI` flags on `FuncDecl`
  - `internal/ast/print.go`: updated pretty-printer for new nodes
  - `internal/parser/parser.go`: `async func`, `ui func`, `await` expressions, `Task<T>` generic type annotations
  - `internal/checker/types.go`: `Async`/`UI` fields on `FuncSignature`
  - `internal/checker/builtins.go`: `Task` generic type sentinel registration
  - `internal/checker/checker.go`: async-safety enforcement (E060-E064), `await` expression checking, `Task<T>` return wrapping for async calls
  - `internal/interpreter/value.go`: `TaskVal` with channel-based signaling (`newTaskVal`, `resolve`, `await`), `Async`/`UI` on `FuncVal`
  - `internal/interpreter/interpreter.go`: async func spawns goroutine and returns `TaskVal` immediately, `await` blocks until task completes

### Fixed

- `internal/checker/checker.go`: `asyncPayloadType` now checks `RecordType.Name == "Error"` instead of matching any `RecordType` when stripping trailing `Error?` from async return tuples

## [0.0.1] - 30-05-2026

- Added `docs/superpowers/specs/2026-05-31-threading-concurrency-design.md` documenting the proposed first threading/concurrency model: per-window UI actors, window-owned `ui func`s, `async func` tasks, `await`, sendable task boundaries, and cancel-on-close lifecycle semantics.

## [0.0.1] - 28-05-2026

### Fixed

- Constant reassignment now produces a runtime error (E101) instead of silently shadowing the binding (`internal/interpreter/env.go`, `internal/interpreter/interpreter.go`)
- Exclusive range with equal bounds (e.g. `0..<0`) now correctly produces an empty sequence instead of `[0, -1]` (`internal/interpreter/interpreter.go`)

### Added

- Go compiler Phase 4 tree-walking interpreter implementation:
  - `internal/interpreter` package: runtime values, lexical environments, control-flow signals, expression/statement evaluator, function calls, records, constructors, methods, loops, switches, interpolation, lists, dicts, and ranges
  - Runtime built-in namespaces for `console`, `file`, `json`, and `system`, plus UI-related stub namespaces matching the checker
  - Runtime constant groups for colors, dock, orientation, toast types, font weights, text alignment, scroll modes, and sort order
  - Runtime diagnostic codes E100-E107 for undefined variables, type mismatches, indexing, division by zero, non-callable values, argument counts, dict key errors, and call depth
  - `leal run <file.ll> [args...]` CLI command that lexes, parses, type-checks, then executes programs
  - Interpreter tests covering console output, arithmetic, control flow, loops, functions, recursion, records, methods, interpolation, collections, builtins, runtime errors, and `examples/full.ll`

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

### Fixed

- Type checker: removed double type-checking of call arguments in `checkCallArgs` (duplicate diagnostics)
- Type checker: removed double type-checking of loop iterator iterables in `checkLoopStmt` (duplicate diagnostics)
- Type checker: added specific E039 error for missing record methods instead of misleading E049 "cannot call non-function type"
- Diagnostics: checker and parser errors now display source line excerpts (added `SetSource` fallback to `Diagnostics`)
- Type checker: loop iterables now always type-checked even when using `_` variable (e.g., `for _ in some_func():`), with regression test
- Type checker: assignment error span now correctly highlights the value expression instead of starting at the `=` sign, with regression test
- Diagnostics: error highlighting now spans the full token (e.g., `~~~~` for `true`) instead of a single character
- Parser: `errorAt` now computes highlight span from token lexeme length

## [0.0.1] - 27-05-2006

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
