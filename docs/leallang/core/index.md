# Core Language

## Design Goals

LealLang should be:

1. **Declarative for UI**

   Window and component trees should be readable as a visual hierarchy.

2. **Explicit for behavior**

   Event handlers, errors, UI references, and mutations should be visible in code.

3. **Typed**

   Variables, records, function parameters, function returns, component properties, and event handlers should be type-checked.

4. **Small but practical**

   The language should be powerful enough to build desktop GUI applications without trying to become a general-purpose replacement for Go, Python, or C#.

5. **Predictable**

   Assignment, copying, nullability, and mutation rules should be explicit.

6. **Package-oriented**

   Applications should be split into packages with private-by-default symbols and explicit exports.

---

## Reference Pages

- [Lexical Structure](lexical.md)
- [Expressions and Operators](expressions.md)
- [Packages and Imports](packages.md)
- [Types and Data](types.md)

---

## Non-Goals

LealLang does not aim to be:

- A general-purpose systems language
- An object-oriented language
- A class-based inheritance language
- A Python clone
- A C# clone
- A JavaScript replacement
- A web framework language
- A language with implicit exception propagation

LealLang is primarily a desktop GUI language.
