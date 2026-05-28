# Lexical Structure

LealLang uses indentation-based scoping and colon-delimited blocks.

```python
func greet(name: string):
    console.print_ln($"Hello, {name}")
```

Blocks are introduced with `:`.

```python
if is_ready:
    console.print_ln("Ready")
else:
    console.print_ln("Not ready")
```

Parentheses are required for function calls and component declarations.

```python
console.print_ln("Hello")
Button[save_button]("Save")
```

Names are case-sensitive.

```python
name
Name
NAME
```

These are three different identifiers.

---

## Comments

Single-line comments use `#`.

```python
# This is a single-line comment
```

Multi-line comments use `/* */`.

```python
/*
  This is a multi-line comment.
  It can span multiple lines.
*/
```

---

## Empty Statement

`pass` is an empty statement used where a block is required but no action is needed.

```python
func todo():
    pass
```

---

## Bracketed Syntax Forms

LealLang uses brackets in several distinct syntax forms.

Component declarations create UI nodes and do not use `@`.

```python
Button[save_button]("Save")
```

UI references access existing UI nodes and use `@`.

```python
@Button[save_button]
@Window[settings].Button[close_button]
```

Component IDs such as `save_button` and `settings` are not normal variables. They are identifiers inside the UI tree and are accessed through UI references.

Collection indexing uses `[]` on values.

```python
numbers[0]
scores["Ana"]
```

Generic types use angle brackets.

```python
List<string>
Dict<string, int>
```

---

## Discard and Wildcard `_`

`_` discards an assigned value.

```python
_, err = file.read_text("./settings.json")
value, _ = cache.get("theme")
```

Values assigned to `_` cannot be read later.

`_` is also the wildcard/default arm in switch expressions and switch statements.

```python
color = switch color_name:
    "blue" -> colors.blue
    _ -> colors.white
```

---

## Naming Conventions

LealLang uses `snake_case` consistently.

Variables:

```python
user_name = "Ana"
current_count = 10
```

Functions:

```python
func load_config(path: string):
    pass
```

Record fields:

```python
type User:
    pub first_name: string
    pub last_name: string
```

Built-in namespaces and functions:

```python
console.print_ln("Hello")
file.read_text("./config.json")
window.open(@Window[settings])
```

Built-in constants:

```python
colors.white
dock.fill
orientation.horizontal
```

Window and component IDs use `snake_case`.

```python
Window[main_window]("My App"):
    Button[save_button]("Save")
```
