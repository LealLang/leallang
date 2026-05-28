# Expressions and Operators

Expressions produce values. They are used in assignments, function calls, component arguments, conditions, loop headers, and return statements.

## Literals

```python
1
1.5
true
false
null
"text"
'c'
```

`null` can only be used with nullable types such as `string?` or `Error?`.

## Names and Field Access

Names refer to variables, constants, functions, imported packages, and type names.

```python
user_name
```

Fields are accessed with `.`.

```python
user.name
err.message
```

## Function Calls

Function calls use parentheses.

```python
load_config(path)
console.print_ln("Saved")
```

Named arguments use `name: value`.

```python
load_config(path: "./config.json")
toast.show(message: "Saved", duration: 3)
```

Positional arguments must come before named arguments. After the first named argument, all remaining arguments must be named.

```python
toast.show("Saved", duration: 3)        # valid
toast.show(message: "Saved", duration: 3) # valid
toast.show(message: "Saved", 3)        # invalid
```

Component declarations have stricter positional argument rules. See [Components](../components/index.md).

## Generic Types and Calls

Generic types use angle brackets.

```python
List<string>
Dict<string, int>
```

Generic calls place type arguments before the call arguments.

```python
config, err = json.parse<Config>(text)
```

## Indexing

Lists and dictionaries use `[]` for indexing.

```python
first = numbers[0]
score = scores["Ana"]
```

Direct dictionary indexing assumes the key exists.

```python
score = scores["Ana"]
```

If the key is missing, direct access fails at runtime.

Safe dictionary access uses `.get`.

```python
score, err = scores.get("Ana")
if err != null:
    console.print_ln("Missing score")
```

## UI References

UI references use `@` and refer to windows or components in the UI tree.

```python
@Button[save_button]
@Window[settings].Button[close_button]
```

Component IDs are not normal variables. They are identifiers inside the UI tree and are accessed through `@` references.

## Component Declarations

Component declarations do not use `@`.

```python
Button[save_button]("Save")
Panel[side_panel](bg: colors.white, dock: dock.left):
    Label[title_label]("Settings")
```

A declaration creates a window or component in the UI tree. A UI reference accesses an existing window or component.

## Record Construction

Records use named-field construction.

```python
config = Config(theme: "light", autosave: false)
```

Positional record construction is invalid.

```python
config = Config("light", false) # invalid
```

## Collection Literals

List literals use `[]`.

```python
numbers: List<int> = [1, 2, 3]
```

Dictionary literals use key-value entries.

```python
scores: Dict<string, int> = {
    "Ana": 10,
    "Bob": 8
}
```

## String Interpolation

String interpolation uses `$"..."`.

```python
message = $"Hello, {name}"
console.print_ln($"Total: {price * quantity}")
```

## Operators

Arithmetic operators:

```text
+  -  *  /  %
```

Comparison operators:

```text
==  !=  <  <=  >  >=
```

Boolean operators:

```text
and  or  not
```

Examples:

```python
total = price * quantity
is_valid = count > 0 and err == null
if not is_ready:
    console.print_ln("Waiting")
```

## Ranges

Inclusive ranges use `..`.

```python
0..3 # 0, 1, 2, 3
```

Exclusive upper-bound ranges use `..<`.

```python
0..<3 # 0, 1, 2
```

Exclusive ranges are useful for collection indexing.

```python
loop i in 0..<items.count():
    console.print_ln(items[i])
```

## Switch Expressions

Switch expressions return values. Expression arms use `->`.

```python
color = switch color_name:
    "blue" -> colors.blue
    "red" -> colors.red
    "purple" | "violet" -> colors.purple
    _ -> colors.white
```

Statement switch arms use `:`. See [Control Flow](../functions/control-flow.md).

## `any` Checks and Extraction

Values of type `any` require explicit narrowing or checked extraction before assignment to a concrete type.

```python
value: any = "hello"
value = 42
```

Narrowing with `is`:

```python
if value is string:
    name: string = value
    console.print_ln(name)
```

Checked extraction with `.as<T>()`:

```python
number, ok = value.as<int>()
if ok:
    console.print_ln(number)
```

Silent assignment from `any` to a concrete type is invalid.

```python
name: string = value # invalid
```
