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

For some concrete examples, look at files in the `testdata` directory.

## Usage

[![Go Reference](https://pkg.go.dev/badge/github.com/eliben/go-ungrammar.svg)](https://pkg.go.dev/github.com/eliben/go-ungrammar)

Usage example:

https://github.com/eliben/go-ungrammar/blob/229d0dd20660980d5069ed676c5c728a9fda5723/example_test.go#L13-L31

For somewhat more sophisticated usage, see the `cmd/ungrammar2json` command.
