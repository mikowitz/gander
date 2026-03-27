# Gander: A Zipper Library for Go

**Module:** `github.com/mikowitz/gander`
**Go version:** 1.23+
**Based on:** Huet's Zipper (1997) and Clojure's `clojure.zip`

---

## 1. Design Principles

- **Immutable:** All operations return new values; nothing is mutated in place.
- **Interface-driven:** Users implement `Node` (and either `Leaf` or `Branch`) on their own types to get a zipper for free.
- **Idiomatic Go:** Navigation returns `(Zipper, bool)` — `false` means "can't go there" (at top, at leftmost, no children, etc.). No panics for normal navigation failures. Edit operations that are structurally invalid (e.g., `InsertLeft` at root) also return `(Zipper, bool)`.
- **Simplicity over deep-tree performance:** `path.pnodes` stores the full ancestor list at every level as a plain Go slice. Because slices must be copied on each navigation step (no structural sharing), navigating to depth N allocates O(N²) total memory across the path chain. Clojure's persistent vectors avoid this cost; Go does not have an equivalent. This is a conscious trade-off: `Path()` is O(1) with this design, and real-world trees (ASTs, file systems, XML) are rarely deeper than a few dozen levels, so the quadratic cost is unlikely to be observable in practice.

---

## 2. Core Types

### 2.1 Node Interfaces

```go
package gander

// Node is the base interface for all tree nodes.
// The unexported marker method prevents accidental implementation.
type Node interface {
    node()
}

// Leaf is a terminal node with no children.
type Leaf interface {
    Node
}

// Branch is a node that can have children (even if currently empty).
type Branch interface {
    Node
    Children() []Node
    MakeNode(children []Node) Node
}
```

**Key distinction:** `Branch` means "can have children," not "has children right now." An empty directory or an empty slice should still implement `Branch`.

The `node()` unexported marker prevents arbitrary types from satisfying `Node`. Users embed a helper or implement it explicitly:

```go
// Convenience type users can embed
type BaseNode struct{}
func (BaseNode) node() {}
```

### 2.2 Path (unexported)

Mirrors the Clojure implementation's path map. This is internal to the package.

```go
// path represents the context above the current focus.
// nil path means we are at the root.
type path struct {
    left    []Node  // left siblings, rightmost-first (stack order, like Clojure's :l)
    parent  *path   // parent path (nil if parent is root-level)
    pnodes  []Node  // ancestor nodes (stack, like Clojure's :pnodes)
    right   []Node  // right siblings, leftmost-first
    changed bool    // whether edits have occurred below
}
```

**`pnodes` vs `parent`:** These two fields serve distinct roles and are both necessary.

- `pnodes` is a growing slice of ancestor *nodes* from root down to the immediate parent. `pnodes[len(pnodes)-1]` is the parent node itself — the value passed to `MakeNode` during `Up` when reconstructing the tree.
- `parent` is the *context* (path) that was active at the parent's level — it records the parent's own left/right siblings, its grandparent path, and its own changed flag. It is the context needed to continue navigating upward past the parent.

Together: `pnodes[last]` answers "what is the parent node?"; `parent` answers "what was the surrounding context when we entered the parent?"

**Example:** given `root → [A → [X, Y], B, C]` with focus on `Y`:
```
loc.focus  = Y
loc.path   = &path{
    left:    [X],               // X is left of Y (nearest first)
    right:   [],                // Y is rightmost
    pnodes:  [root, A],         // ancestors root→A
    parent:  &path{             // context that was active when we were at A
        left:    [],
        right:   [B, C],
        pnodes:  [root],
        parent:  nil,           // root has no parent context
        changed: false,
    },
    changed: false,
}
```

When `Up` is called from `Y`, it uses `pnodes[last]` (which is `A`) to call `A.MakeNode([X, Y])`, then returns a `Zipper` with the reconstructed `A` as focus and `parent` as the new path.

### 2.3 Zipper

```go
// Zipper is a location in the zipper: a focused node plus its surrounding context.
// The zero value is not valid; create via New().
type Zipper struct {
    focus Node
    path  *path  // nil means root with no navigation history
    end   bool   // true if this is the end sentinel from depth-first traversal
}
```

---

## 3. Constructor

```go
// New creates a zipper rooted at the given node.
func New(root Node) Zipper
```

Returns `Zipper{focus: root, path: nil}`.

---

## 4. API Surface

All functions are package-level, taking `Zipper` as first argument (value receiver style, since `Zipper` is immutable).

### 4.1 Accessors

| Function | Signature | Description |
|----------|-----------|-------------|
| `Focus` | `(Zipper) Node` | Returns the node at the current focus. |
| `IsBranch` | `(Zipper) bool` | True if the focused node implements `Branch`. |
| `Children` | `(Zipper) ([]Node, bool)` | Children of the focused node. False if not a branch. |
| `Path` | `(Zipper) []Node` | Ancestor nodes from root down to (but not including) current focus. |
| `Lefts` | `(Zipper) []Node` | Left siblings in tree order (leftmost first). Note: internally stored reversed; return in natural order. |
| `Rights` | `(Zipper) []Node` | Right siblings in tree order. |

### 4.2 Navigation

All return `(Zipper, bool)` where `false` means the move is not possible.

| Function | Signature | Description |
|----------|-----------|-------------|
| `Down` | `(Zipper) (Zipper, bool)` | Move to the leftmost child. False if leaf or empty branch. |
| `Up` | `(Zipper) (Zipper, bool)` | Move to parent, reconstructing it if changes were made below. False if at root. |
| `Left` | `(Zipper) (Zipper, bool)` | Move to left sibling. False if leftmost or at root. |
| `Right` | `(Zipper) (Zipper, bool)` | Move to right sibling. False if rightmost or at root. |
| `Leftmost` | `(Zipper) (Zipper, bool)` | Move to leftmost sibling. Returns self and true if already there; false at root. |
| `Rightmost` | `(Zipper) (Zipper, bool)` | Move to rightmost sibling. Returns self and true if already there; false at root. |
| `Root` | `(Zipper) (Zipper, bool)` | Zip all the way up, applying changes, returning a Zipper at root. If called on an end sentinel, returns a new root Zipper wrapping the sentinel's focus node directly (with nil path), since the end sentinel's focus already holds the fully-accumulated root with all edits applied. |

### 4.3 Depth-First Traversal

| Function | Signature | Description |
|----------|-----------|-------------|
| `Next` | `(Zipper) Zipper` | Move to the next location in depth-first order. Returns end sentinel when exhausted. |
| `Prev` | `(Zipper) (Zipper, bool)` | Move to the previous location in depth-first order. False if at root. |
| `IsEnd` | `(Zipper) bool` | True if this Zipper is the end sentinel. |

### 4.4 Editing

All editing operations return `(Zipper, bool)`. False if the operation is structurally invalid (e.g., insert at root).

| Function | Signature | Description |
|----------|-----------|-------------|
| `Replace` | `(Zipper, Node) (Zipper, bool)` | Replace the focused node. False if at end sentinel. |
| `Edit` | `(Zipper, func(Node) Node) (Zipper, bool)` | Apply function to the focused node and replace. False if at end sentinel. |
| `InsertLeft` | `(Zipper, Node) (Zipper, bool)` | Insert a node as the left sibling of focus. False at root. |
| `InsertRight` | `(Zipper, Node) (Zipper, bool)` | Insert a node as the right sibling of focus. False at root. |
| `InsertChild` | `(Zipper, Node) (Zipper, bool)` | Insert as leftmost child of focus. False if not a branch. |
| `AppendChild` | `(Zipper, Node) (Zipper, bool)` | Insert as rightmost child of focus. False if not a branch. |
| `Remove` | `(Zipper) (Zipper, bool)` | Remove focused node, moving to what would have preceded it in depth-first walk. False at root. |

---

## 5. Immutability Contract

- `Zipper` is a value type. All operations return new `Zipper` values.
- `path` is a pointer type for structural sharing, but is never mutated after creation. New paths are always fresh allocations with copied/new slices.
- Slice fields (`left`, `right`, `pnodes`) must be copied on modification, never appended in place (append can mutate shared backing arrays).

---

## 6. Implementation Steps

Each step is a self-contained unit with tests written first (TDD red-green-refactor). Each step should be under ~200 lines of implementation code.

### Step 1: Interfaces and Zipper Type

**Files:** `node.go`, `zipper.go`, `node_test.go`

Implement:
- `Node`, `Leaf`, `Branch` interfaces
- `BaseNode` embed helper
- `Zipper` struct (unexported `path` struct)
- `New(root Node) Zipper`
- `Focus(Zipper) Node`
- `IsBranch(Zipper) bool`
- `Children(Zipper) ([]Node, bool)`

**Test types for this step and all subsequent steps:**

```go
// A simple leaf
type StringLeaf struct {
    gander.BaseNode
    Value string
}

func (s StringLeaf) Equal(other gander.Node) bool {
    o, ok := other.(StringLeaf)
    return ok && s.Value == o.Value
}

// A simple branch
type ListBranch struct {
    gander.BaseNode
    Items []gander.Node
}

func (lb ListBranch) Children() []gander.Node { return lb.Items }
func (lb ListBranch) MakeNode(children []gander.Node) gander.Node {
    return ListBranch{Items: children}
}

// Equal uses reflect.DeepEqual because ListBranch contains a slice field
// and is not directly comparable with ==.
func (lb ListBranch) Equal(other gander.Node) bool {
    return reflect.DeepEqual(lb, other)
}

// CountingBranch is a ListBranch variant that tracks MakeNode calls.
// Use it to verify the changed flag optimization: MakeNode should only
// be called when edits have been made in the subtree below this node.
type CountingBranch struct {
    gander.BaseNode
    Items     []gander.Node
    makeCount *int // pointer so the count is shared across value copies
}

func NewCountingBranch(items []gander.Node) CountingBranch {
    n := 0
    return CountingBranch{Items: items, makeCount: &n}
}

func (cb CountingBranch) Children() []gander.Node { return cb.Items }

func (cb CountingBranch) MakeNode(children []gander.Node) gander.Node {
    *cb.makeCount++
    return CountingBranch{Items: children, makeCount: cb.makeCount}
}

func (cb CountingBranch) MakeCount() int { return *cb.makeCount }

func (cb CountingBranch) Equal(other gander.Node) bool {
    return reflect.DeepEqual(cb, other)
}
```

**Note on equality in tests:** `ListBranch` and `CountingBranch` contain slice fields and are not comparable with `==` — doing so panics at runtime. Always use the `.Equal` method for structural equality assertions. `StringLeaf` supports `==` directly but `.Equal` is preferred for consistency.

**Tests:**
- `New` returns a `Zipper` with the root as focus
- `Focus` returns the focused node
- `IsBranch` returns true for `ListBranch`, false for `StringLeaf`
- `Children` returns children of a branch, `false` for a leaf
- `Children` returns empty slice and `true` for an empty branch

### Step 2: Down and Up Navigation

**Files:** `nav.go`, `nav_test.go`

Implement:
- `Down(Zipper) (Zipper, bool)`
- `Up(Zipper) (Zipper, bool)`

**Tests:**
- `Down` on a branch with children focuses leftmost child
- `Down` on a leaf returns false
- `Down` on an empty branch returns false
- `Up` from a child returns to the parent
- `Up` from root returns false
- `Down` then `Up` round-trips: `Focus` returns a node `Equal` to the original
- `Down` then `Up` with no edits does not call `MakeNode` (use `CountingBranch`, assert `MakeCount() == 0`)
- `Down` sets correct `right` siblings
- Nested `Down` → `Down` → `Up` → `Up` works

### Step 3: Left and Right Navigation

**Files:** Add to `nav.go`, `nav_test.go`

Implement:
- `Left(Zipper) (Zipper, bool)`
- `Right(Zipper) (Zipper, bool)`
- `Leftmost(Zipper) (Zipper, bool)`
- `Rightmost(Zipper) (Zipper, bool)`

**Tests:**
- `Down` then `Right` focuses second child
- `Right` at rightmost returns false
- `Left` at leftmost returns false
- `Left`/`Right` at root returns false
- `Right` then `Left` round-trips
- `Leftmost` from any sibling goes to first
- `Rightmost` from any sibling goes to last
- `Leftmost`/`Rightmost` at root returns false

### Step 4: Path, Lefts, Rights Accessors

**Files:** `accessors.go`, `accessors_test.go`

Implement:
- `Path(Zipper) []Node`
- `Lefts(Zipper) []Node`
- `Rights(Zipper) []Node`

**Tests:**
- `Path` at root returns empty slice
- `Path` after `Down` returns `[root]`
- `Path` after `Down` → `Down` returns `[root, firstChild]`
- `Lefts` after `Down` returns empty (leftmost child)
- `Lefts` after `Down` → `Right` returns `[firstChild]`
- `Rights` after `Down` returns remaining siblings
- `Rights` after `Rightmost` returns empty

### Step 5: Replace and Edit

**Files:** `edit.go`, `edit_test.go`

Implement:
- `Replace(Zipper, Node) (Zipper, bool)`
- `Edit(Zipper, func(Node) Node) (Zipper, bool)`

**Tests:**
- `Replace` changes the focused node, `Focus` reflects it
- `Replace` then `Up` then `Focus` shows reconstructed parent with new child
- `Replace` then `Root` returns modified tree
- `Edit` applies function to current node
- `Replace` deep in tree, then `Root` propagates changes all the way up
- Unmodified `Up` does NOT reconstruct: use `CountingBranch` as the parent, navigate into a child and back up without editing — assert `MakeCount() == 0`
- Modified `Up` DOES reconstruct: same setup, call `Replace` on the child before `Up` — assert `MakeCount() == 1`

### Step 6: Insert Left, Insert Right

**Files:** Add to `edit.go`, `edit_test.go`

Implement:
- `InsertLeft(Zipper, Node) (Zipper, bool)`
- `InsertRight(Zipper, Node) (Zipper, bool)`

**Tests:**
- `InsertLeft` at root returns false
- `InsertRight` at root returns false
- `InsertLeft` adds to left siblings, focus stays
- `InsertRight` adds to right siblings, focus stays
- Insert then `Root` shows the new sibling in the tree
- Insert then navigate left/right finds the inserted node
- Multiple inserts accumulate correctly

### Step 7: Insert Child, Append Child

**Files:** Add to `edit.go`, `edit_test.go`

Implement:
- `InsertChild(Zipper, Node) (Zipper, bool)`
- `AppendChild(Zipper, Node) (Zipper, bool)`

**Tests:**
- `InsertChild` on a leaf returns false
- `AppendChild` on a leaf returns false
- `InsertChild` adds as leftmost child, focus stays on parent
- `AppendChild` adds as rightmost child, focus stays on parent
- `InsertChild` then `Down` focuses the inserted child
- `AppendChild` on empty branch creates a single child
- Changes propagate to `Root`

### Step 8: Remove

**Files:** Add to `edit.go`, `edit_test.go`

Implement:
- `Remove(Zipper) (Zipper, bool)`

**Algorithm sketch:**

```
Remove(loc):
  if loc.path == nil:
    return _, false   // at root

  if len(loc.path.left) > 0:
    // Predecessor is the rightmost descendant of the nearest left sibling.
    // Drop the current focus; move to that left sibling.
    pred := Zipper{
      focus: loc.path.left[0],
      path: &path{
        left:    loc.path.left[1:],   // remaining lefts
        parent:  loc.path.parent,
        pnodes:  loc.path.pnodes,
        right:   loc.path.right,      // current node is gone; its rights remain
        changed: true,
      },
    }
    // Descend to rightmost descendant of pred
    for IsBranch(pred) && len(Children(pred)) > 0:
      pred, _ = Down(pred)
      pred, _ = Rightmost(pred)
    return pred, true

  else:
    // Current node is the leftmost child.
    // Reconstruct parent with only the right siblings as its new children.
    newParent := loc.path.pnodes[last].MakeNode(loc.path.right)
    parentPath := loc.path.parent  // nil when grandparent is root
    if parentPath != nil:
      parentPath = &path{...parentPath, changed: true}
    return Zipper{focus: newParent, path: parentPath}, true
```

**Tests:**
- `Remove` at root returns false
- `Remove` with left siblings moves to depth-first predecessor (rightmost descendant of left sibling)
- `Remove` at leftmost child moves to parent (with remaining children)
- `Remove` only child leaves parent with empty children
- `Remove` then `Root` shows the node is gone
- `Remove` then `Next` continues traversal correctly

### Step 9: Depth-First Traversal (Next, Prev, IsEnd)

**Files:** `traverse.go`, `traverse_test.go`

Implement:
- `Next(Zipper) Zipper`
- `Prev(Zipper) (Zipper, bool)`
- `IsEnd(Zipper) bool`

**Tests:**
- `Next` from root goes to first child (if branch)
- `Next` visits all nodes in depth-first order
- `Next` after last node returns end sentinel
- `IsEnd` on end sentinel returns true
- `IsEnd` on normal zipper returns false
- `Next` on end sentinel returns end sentinel (stays)
- `Root` on end sentinel returns false
- `Prev` reverses `Next`
- `Prev` at root returns false
- Full traversal: collect all nodes via `Next` loop, verify order
- Traversal with edits: `Next` + conditional `Replace`, then `Root` yields modified tree

### Step 10: Immutability Verification and Integration Tests

**Files:** `integration_test.go`

No new implementation — this step is purely tests verifying the full contract:

- **Immutability:** Performing operations on a `Zipper` does not affect any previously-held `Zipper` values. Navigate to a child, hold that `Zipper`, navigate further and make edits, then verify the held `Zipper`'s `Focus` is still `Equal` to what it was when captured.
- **Round-trip:** Build a tree, traverse the entire thing with `Next` without editing, call `Root` — verify the result is `Equal` to the original root. Additionally use `CountingBranch` as the root to assert `MakeCount() == 0`, confirming no reconstruction occurred.
- **Structural sharing:** In a tree of depth N, make one edit at the deepest level, call `Root`, and verify via `CountingBranch` that `MakeNode` was called exactly N times (once per ancestor), not more.
- **Complex edit scenario:** Replicate the Clojure example: `[[a * b] + [c * d]]` → replace all `*` with `/` via `Next` loop → verify result.
- **Remove during traversal:** Remove nodes during a `Next` walk, verify `Root` result.
- **Heterogeneous tree:** Build a tree with multiple concrete `Branch` and `Leaf` types mixed together, navigate and edit it.
- **Empty branch handling:** Navigate into and out of empty branches.

---

## 7. File Layout

```
github.com/mikowitz/gander/
├── go.mod
├── node.go          // Node, Leaf, Branch interfaces, BaseNode
├── zipper.go        // Zipper, path structs, New(), Focus(), IsBranch(), Children()
├── nav.go           // Down, Up, Left, Right, Leftmost, Rightmost, Root
├── accessors.go     // Path, Lefts, Rights
├── edit.go          // Replace, Edit, InsertLeft, InsertRight, InsertChild, AppendChild, Remove
├── traverse.go      // Next, Prev, IsEnd
├── node_test.go     // Test helpers and type tests
├── nav_test.go      // Navigation tests
├── accessors_test.go
├── edit_test.go
├── traverse_test.go
└── integration_test.go
```

---

## 8. Implementation Notes

### Slice Copying

When building new `path` values, always copy slices rather than appending:

```go
// WRONG — may mutate shared backing array
newLeft := append(p.left, node)

// CORRECT — fresh slice
newLeft := make([]Node, len(p.left)+1)
copy(newLeft, p.left)
newLeft[len(p.left)] = node
```

### Left Sibling Storage Order

Like Clojure's `:l`, left siblings are stored in stack order (rightmost-nearest-sibling first). The `Lefts()` accessor must reverse this to return natural tree order.

### Change Propagation

The `changed` flag on `path` controls whether `Up` reconstructs the parent via `MakeNode` or returns the original parent node from `pnodes`. This is a structural-sharing optimization: unmodified subtrees are never copied.

**When `changed` is set to true:**
Any edit operation (`Replace`, `Edit`, `InsertLeft`, `InsertRight`, `InsertChild`, `AppendChild`, `Remove`) must set `changed: true` on the path of the resulting `Zipper`.

**How `Up` uses it:**

```
Up(loc):
  if loc.path == nil:
    return _, false

  p        := loc.path
  parentNode := p.pnodes[last]

  if p.changed:
    // Rebuild parent: lefts (reversed to natural order) + focus + rights
    children   = reverse(p.left) + [loc.focus] + p.right
    parentNode = parentNode.MakeNode(children)
    // Propagate changed upward to grandparent's path
    grandparentPath = p.parent
    if grandparentPath != nil:
      grandparentPath = &path{...grandparentPath, changed: true}
  else:
    grandparentPath = p.parent   // unchanged; no reconstruction needed

  return Zipper{focus: parentNode, path: grandparentPath}, true
```

**What this means for callers:** After a pure navigation round-trip (no edits), the node returned by `Focus` is the identical original node — not a freshly constructed copy. After any edit, every ancestor on the path back to root is reconstructed exactly once when `Up` or `Root` is called.

**Testing the optimization with `CountingBranch`:**

Verify `MakeNode` is NOT called during unmodified navigation:

```go
cb  := NewCountingBranch([]gander.Node{StringLeaf{"a"}, StringLeaf{"b"}})
loc := gander.New(cb)
loc, _ = gander.Down(loc)
loc, _ = gander.Up(loc)
// No edit occurred — MakeNode must not have been called
assert(cb.MakeCount() == 0)
```

Verify `MakeNode` IS called exactly once after an edit and `Up`:

```go
cb  := NewCountingBranch([]gander.Node{StringLeaf{"a"}, StringLeaf{"b"}})
loc := gander.New(cb)
loc, _ = gander.Down(loc)
loc, _ = gander.Replace(loc, StringLeaf{"z"})
loc, _ = gander.Up(loc)
assert(cb.MakeCount() == 1)
```

For a deeper tree, verify the count equals the number of ancestors between the edit and the root — each ancestor is reconstructed at most once.

### Prev Algorithm

`Prev` is the depth-first inverse of `Next`. Three cases:

```
Prev(loc):
  if loc.path == nil:
    return _, false   // at root — no predecessor

  if len(loc.path.left) > 0:
    // There is a left sibling. The predecessor is that sibling's
    // rightmost descendant (or the sibling itself if it is a leaf).
    sibling := Zipper{
      focus: loc.path.left[0],
      path: &path{
        left:    loc.path.left[1:],
        parent:  loc.path.parent,
        pnodes:  loc.path.pnodes,
        right:   prepend(loc.focus, loc.path.right),  // [focus] + rights
        changed: loc.path.changed,
      },
    }
    // Descend to rightmost descendant of sibling
    for IsBranch(sibling) && len(Children(sibling)) > 0:
      sibling, _ = Down(sibling)
      sibling, _ = Rightmost(sibling)
    return sibling, true

  else:
    // No left siblings — parent comes immediately before all its children
    // in depth-first order, so the parent is the predecessor.
    return Up(loc)
```

**`prepend` note:** `right` is stored in natural (leftmost-first) order, so prepending `loc.focus` means constructing a new slice `[loc.focus] + loc.path.right...`.

This mirrors the predecessor-finding logic in `Remove` but keeps the removed node in the tree; it's purely a navigation move.

### End Sentinel

`IsEnd` checks the `end` field on `Zipper`. `Next` on an end sentinel returns itself.

`Root` on an end sentinel does **not** return false. By the time `Next` produces an end sentinel, it has zipped all the way up to root (the traversal algorithm must travel up repeatedly until `Up` fails, which only happens at root), so the sentinel's `focus` is already the fully-accumulated root node (with all edits applied) and `path` is nil. `Root` on an end sentinel therefore returns `Zipper{focus: sentinel.focus, path: nil}` with `true` — a fresh, non-sentinel root Zipper.

This makes the common traversal-with-edits pattern work naturally:

```go
loc := gander.New(root)
for !gander.IsEnd(loc) {
    if condition(gander.Focus(loc)) {
        loc, _ = gander.Replace(loc, newNode)
    }
    loc = gander.Next(loc)
}
result, _ := gander.Root(loc) // works: returns the modified root
```

The Step 9 test "Root on end sentinel returns false" should be updated to: `Root` on end sentinel returns a valid root Zipper (true), not false.
