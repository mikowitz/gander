# gander

A tree zipper library for Go.

A zipper is a functional, immutable cursor for navigating and editing arbitrary
tree structures. It was introduced by Gérard Huet in his 1997 paper
[The Zipper](https://www.st.cs.uni-saarland.de/edu/seminare/2005/advanced-fp/docs/huet-zipper.pdf)
and popularized in Clojure by `clojure.zip`. Gander closely follows the
`clojure.zip` interface.

## Installation

```sh
go get github.com/mikowitz/gander
```

Requires Go 1.25 or later.

## Overview

A [`Zipper`](https://pkg.go.dev/github.com/mikowitz/gander#Zipper) is a value
that holds a *focused* node and the surrounding context needed to navigate back
to any ancestor or sibling. Because `Zipper` is a value type, all operations
return new values — the original is never modified.

## Defining Tree Nodes

Implement the [`Node`](https://pkg.go.dev/github.com/mikowitz/gander#Node)
interface on your own types. Embed
[`BaseNode`](https://pkg.go.dev/github.com/mikowitz/gander#BaseNode) to satisfy
the unexported marker method. Types that can have children must also implement
[`Branch`](https://pkg.go.dev/github.com/mikowitz/gander#Branch).

```go
// A leaf node
type StringLeaf struct {
    gander.BaseNode
    Value string
}

// A branch node
type ListBranch struct {
    gander.BaseNode
    Items []gander.Node
}

func (b ListBranch) Children() []gander.Node { return b.Items }
func (b ListBranch) WithChildren(children []gander.Node) gander.Branch {
    return ListBranch{Items: children}
}
```

**`Branch` means "can have children,"** not "currently has children." An empty
node (empty directory, empty list) should still implement `Branch`.

## Basic Usage

```go
a := StringLeaf{Value: "a"}
b := StringLeaf{Value: "b"}
c := StringLeaf{Value: "c"}
root := ListBranch{Items: []gander.Node{a, b, c}}

z := gander.NewZipper(root)

// Navigate to the second child
z, _ = gander.Down(z)   // focus: a
z, _ = gander.Right(z)  // focus: b

// Replace the focused node
z, _ = gander.Replace(z, StringLeaf{Value: "B"})

// Walk back to root — changes are propagated automatically
rootZ, _ := gander.Root(z)
// gander.Focus(rootZ) is ListBranch{Items: [a, B, c]}
```

## Depth-First Traversal

Use `Next` and `IsEnd` to walk the entire tree. Edits made during traversal are
accumulated and available via `Root` after the loop.

```go
z := gander.NewZipper(root)

for !gander.IsEnd(z) {
    node := gander.Focus(z)
    if leaf, ok := node.(StringLeaf); ok && leaf.Value == "*" {
        z, _ = gander.Replace(z, StringLeaf{Value: "/"})
    }
    z = gander.Next(z)
}

rootZ, _ := gander.Root(z)
```

## API Reference

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/mikowitz/gander).

### Constructors

| Function | Description |
|----------|-------------|
| `NewZipper(root Node) Zipper` | Create a Zipper focused on the root node. |

### Accessors

| Function | Description |
|----------|-------------|
| `Focus(z) Node` | Return the focused node. |
| `IsBranch(z) bool` | Report whether the focused node implements Branch. |
| `Children(z) ([]Node, bool)` | Return the focused node's children; false if not a Branch. |
| `Path(z) []Node` | Return ancestor nodes from root down to the immediate parent. |
| `Lefts(z) []Node` | Return left siblings in tree order. |
| `Rights(z) []Node` | Return right siblings in tree order. |
| `IsEnd(z) bool` | Report whether z is the depth-first traversal end sentinel. |

### Navigation

All navigation functions return `(Zipper, bool)` where false means the move is
not possible from the current position.

| Function | Description |
|----------|-------------|
| `Down(z) (Zipper, bool)` | Move to the leftmost child. |
| `Up(z) (Zipper, bool)` | Move to the parent, reconstructing it if edits were made below. |
| `Left(z) (Zipper, bool)` | Move to the left sibling. |
| `Right(z) (Zipper, bool)` | Move to the right sibling. |
| `Leftmost(z) (Zipper, bool)` | Move to the leftmost sibling. |
| `Rightmost(z) (Zipper, bool)` | Move to the rightmost sibling. |
| `Root(z) (Zipper, bool)` | Move to the root, applying all pending changes. Always returns true. |

### Traversal

| Function | Description |
|----------|-------------|
| `Next(z) Zipper` | Move to the next node in depth-first order; returns end sentinel when exhausted. |
| `Prev(z) (Zipper, bool)` | Move to the previous node in depth-first order; false at root or end sentinel. |

### Editing

All edit functions return `(Zipper, bool)` where false means the operation is
not valid at the current position.

| Function | Description |
|----------|-------------|
| `Replace(z, n) (Zipper, bool)` | Replace the focused node with n. Always succeeds. |
| `Edit(z, f) (Zipper, bool)` | Apply f to the focused node. Always succeeds. |
| `InsertLeft(z, n) (Zipper, bool)` | Insert n as the left sibling of focus; false at root. |
| `InsertRight(z, n) (Zipper, bool)` | Insert n as the right sibling of focus; false at root. |
| `InsertChild(z, n) (Zipper, bool)` | Insert n as the leftmost child of focus; false if not a Branch. |
| `AppendChild(z, n) (Zipper, bool)` | Append n as the rightmost child of focus; false if not a Branch. |
| `Remove(z) (Zipper, bool)` | Remove the focused node and move to its depth-first predecessor; false at root. |

## Design Notes

**Immutability.** `Zipper` is a value type. Holding a `Zipper` before a series
of operations and reading `Focus` from it afterwards always returns the same
node — subsequent operations on a derived `Zipper` never affect held values.

**Structural sharing.** The `changed` flag on the internal path context ensures
that `Branch.WithChildren` is called only for ancestors of an edited node.
A read-only traversal never triggers any reconstruction.

**Memory.** The ancestor path is stored as a plain slice and copied on each
`Down` step, giving O(N²) total allocation for a traversal to depth N. This is
a conscious trade-off for simplicity; trees deeper than a few hundred levels are
uncommon in practice.
