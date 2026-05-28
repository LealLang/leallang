# CLAUDE.md

## Project Overview

LealLang is a **language design specification** — there is no implementation code, build system, or tests yet. The repository contains only Markdown documentation defining a typed, declarative programming language for building desktop GUI applications.

## Documentation Structure

All spec content lives under `docs/leallang/`. The entry point is `docs/leallang/README.md`.

| Section       | Path          | Covers                                                                                |
| ------------- | ------------- | ------------------------------------------------------------------------------------- |
| Core Language | `core/`       | Design goals, lexical structure, expressions, packages, types                         |
| Functions     | `functions/`  | Function declarations, returns, error handling, control flow                          |
| Components    | `components/` | Windows, widgets, containers, complex components, events                              |
| Built-ins     | `builtins/`   | `console`, `file`, `window`, `msg`, `json`, `modal`, `toast` namespaces and constants |
| Plugins       | `plugins/`    | `.llpkg` plugin system (Lua/HTML/CSS bundles)                                         |
| Examples      | `examples/`   | Application structure (`leal.toml`), full multi-file example                          |

## Key Language Concepts

These concepts span multiple spec sections and are essential for understanding the language:

- **Indentation-based scoping** with colon-delimited blocks (Python-like syntax), `.ll` file extension
- **Package system**: one package per file (`package app.settings`), `import`/`pub` visibility, private-by-default
- **Records** (value types, `type` keyword) vs **collections** (`List<T>`, `Dict<K,V>`, reference types). Assignment copies records but shares collection references. Use `clone()` for deep copies.
- **Go-like error handling**: no exceptions/try/catch. Functions return `Error?` as a trailing value. `Error` has `.message` and `.code` fields. Null-check narrowing unwraps nullable types.
- **UI component tree**: declared as indented hierarchy. Component IDs (`Button[save_button]`) are tree identifiers, not variables. `@ComponentType[id]` syntax references existing UI nodes. `ref` parameters pass mutable aliases.
- **Plugin system**: `.llpkg` packages with `manifest.json`, Lua scripts, HTML templates, and CSS

## Conventions

- `snake_case` for all names (packages, functions, variables, components, record fields)
- Private-by-default; `pub` keyword for cross-package exports
- Nullable types use `?` suffix (e.g., `string?`, `Error?`)
- Generic collections: `List<T>`, `Dict<K, V>`

## RTK - Rust Token Killer

Consider @RTK.

Using `rtk` as a prefix for bash commands is mandatory, example:

```bash
rtk git commit -m "commit message"
rtk git branch
rtk npm install
rtk go version
```

Unique exception is when the command gives error, then you have permission to execute the command without `rtk` prefix.

### Error in bash commands

If you get an error while exeuting a bash command, do not insist trying again more than 2 times, investigate what happened.

## Changelog

### Before implementing

Before implementing read `./docs/agent/CHANGELOG.md` to understand last changes.

### After implementing

After implementing update `./docs/agent/CHANGELOG.md` with changes made following the standtards

