# go-ungrammar

Ungrammar implementation and API in Go. Ungrammar is a DSL for
[concrete syntax trees (CST)](https://en.wikipedia.org/wiki/Parse_tree). For some
background on CSTs and how they relate to ASTs,
see [this blog post](https://eli.thegreenplace.net/2009/02/16/abstract-vs-concrete-syntax-trees/).

This implementation is based on the original
[ungrammar crate](https://github.com/rust-analyzer/ungrammar/), also borrowing
some test files from it.

## Ungrammar syntax

The syntax of Ungrammar files is very simple:

```
//           -- comment
Name =       -- non-terminal definition
'ident'      -- token (terminal)
A B          -- sequence
A | B        -- alternation
A*           -- repetition (zero or more)
A?           -- optional (zero or one)
(A B)        -- grouping elements for precedence control
label:A      -- label hint for naming
```

## Usage

TODO: embed from example_test here
