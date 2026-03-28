// ABOUTME: Defines the core Node interface hierarchy, Zipper type, and fundamental accessor functions.

package gander

// Node is the base interface for all tree nodes. The unexported node method
// acts as a sealed marker, ensuring only types that explicitly opt in (or
// embed [BaseNode]) can satisfy the interface.
type Node interface {
	node()
}

// Leaf is a terminal node with no children. Implementing Leaf is optional;
// any [Node] that does not also implement [Branch] is treated as a leaf.
type Leaf interface {
	Node
}

// Branch is a node that can have children, even if currently empty. Implement
// Branch for any node type that contains child nodes.
//
// Children returns the node's current child nodes.
//
// WithChildren returns a new Branch of the same concrete type populated with
// the given children, leaving the receiver unchanged. Gander calls
// WithChildren when navigating upward through edited subtrees to reconstruct
// ancestor nodes.
type Branch interface {
	Node
	Children() []Node
	WithChildren(children []Node) Branch
}

// BaseNode is a zero-size struct that satisfies the unexported [Node] marker
// method. Embed BaseNode in any concrete node type to implement [Node] without
// writing the node method by hand.
type BaseNode struct{}

func (BaseNode) node() {}

type path struct {
	left, right []Node
	parent      *path
	pnodes      []Node
	changed     bool
}

// Zipper represents a focused location within a tree: the current node plus
// the context needed to navigate to any ancestor or sibling. Zipper is a
// value type; all operations return new Zipper values and leave the receiver
// unchanged. The zero value is not valid; use [NewZipper] to create a Zipper.
type Zipper struct {
	focus Node
	path  *path
	end   bool
}

// NewZipper creates a Zipper focused on root with no surrounding context.
// The returned Zipper is positioned at the root of the tree.
func NewZipper(root Node) Zipper {
	return Zipper{focus: root}
}

// Focus returns the node at the current focus of z.
func Focus(z Zipper) Node {
	return z.focus
}

// IsBranch reports whether the focused node of z implements [Branch].
func IsBranch(z Zipper) bool {
	switch z.focus.(type) {
	case Branch:
		return true
	default:
		return false
	}
}

// Children returns the children of the focused node of z.
// It returns nil, false if the focused node does not implement [Branch].
func Children(z Zipper) ([]Node, bool) {
	if IsBranch(z) {
		return z.focus.(Branch).Children(), true
	} else {
		return nil, false
	}
}

// IsEnd reports whether z is the end sentinel produced by a completed
// depth-first traversal. Calling [Next] on the end sentinel returns the
// sentinel unchanged. Calling [Root] on the end sentinel returns a Zipper
// whose focus is the fully-accumulated root node.
func IsEnd(z Zipper) bool {
	return z.end
}
