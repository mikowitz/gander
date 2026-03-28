// ABOUTME: Implements the Path, Lefts, and Rights accessor functions.
// ABOUTME: These functions expose the surrounding context of the focused node in a Zipper.

package gander

import "slices"

// Path returns the ancestor nodes from the root down to, but not including,
// the focused node of z. The first element is the root and the last is the
// immediate parent of the focused node.
//
// Path returns an empty slice when z is at the root.
func Path(z Zipper) []Node {
	if z.path == nil {
		return []Node{}
	}
	ret := make([]Node, len(z.path.pnodes))
	copy(ret, z.path.pnodes)
	slices.Reverse(ret)
	return ret
}

// Lefts returns the left siblings of the focused node in tree order, with the
// leftmost sibling first. It returns nil if z is at the root; navigate into the
// tree with [Down] before calling Lefts.
func Lefts(z Zipper) []Node {
	if z.path == nil {
		return nil
	}
	return z.path.left
}

// Rights returns the right siblings of the focused node in tree order, with
// the nearest right sibling first. It returns nil if z is at the root; navigate
// into the tree with [Down] before calling Rights.
func Rights(z Zipper) []Node {
	if z.path == nil {
		return nil
	}
	return z.path.right
}
