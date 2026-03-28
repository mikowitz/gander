// ABOUTME: Implements the editing functions for modifying a tree through a Zipper.
// ABOUTME: Covers Replace, Edit, InsertLeft, InsertRight, InsertChild, and AppendChild.

package gander

// Replace returns a Zipper with the focused node replaced by n. The
// surrounding context is preserved; call [Root] to propagate the change up to
// the root. Replace always succeeds.
func Replace(z Zipper, n Node) (Zipper, bool) {
	if z.path == nil {
		return NewZipper(n), true
	}
	z.path.changed = true
	return Zipper{
		focus: n,
		path:  z.path,
	}, true
}

// Edit applies f to the focused node and returns a Zipper with the result as
// the new focus. Edit always succeeds.
func Edit(z Zipper, f func(Node) Node) (Zipper, bool) {
	focus := f(z.focus)
	return Replace(z, focus)
}

// InsertLeft inserts n as the left sibling of the focused node. The focus
// remains on the same node after the insertion.
// It returns z, false when z is at the root.
func InsertLeft(z Zipper, n Node) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}
	z.path.left = append(z.path.left, n)
	z.path.changed = true
	return Zipper{
		focus: z.focus,
		path:  z.path,
	}, true
}

// InsertRight inserts n as the right sibling of the focused node. The focus
// remains on the same node after the insertion.
// It returns z, false when z is at the root.
func InsertRight(z Zipper, n Node) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}
	z.path.right = append([]Node{n}, z.path.right...)
	z.path.changed = true
	return Zipper{
		focus: z.focus,
		path:  z.path,
	}, true
}

// InsertChild inserts n as the leftmost child of the focused node. The focus
// remains on the same node after the insertion.
// It returns z, false if the focused node does not implement [Branch].
func InsertChild(z Zipper, n Node) (Zipper, bool) {
	if IsBranch(z) {
		children := append([]Node{n}, z.focus.(Branch).Children()...)
		focus := z.focus.(Branch).WithChildren(children)
		return Zipper{
			focus: focus,
			path:  z.path,
		}, true
	}
	return z, false
}

// AppendChild appends n as the rightmost child of the focused node. The focus
// remains on the same node after the insertion.
// It returns z, false if the focused node does not implement [Branch].
func AppendChild(z Zipper, n Node) (Zipper, bool) {
	if IsBranch(z) {
		children := append(z.focus.(Branch).Children(), n)
		focus := z.focus.(Branch).WithChildren(children)
		return Zipper{
			focus: focus,
			path:  z.path,
		}, true
	}
	return z, false
}
