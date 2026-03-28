// ABOUTME: Implements the navigation functions for moving between nodes in a Zipper.
// ABOUTME: Covers Down, Up, Left, Right, Leftmost, Rightmost, and Root.

package gander

// Down moves to the leftmost child of the focused node.
// It returns z, false if the focused node is not a [Branch] or has no children.
func Down(z Zipper) (Zipper, bool) {
	if IsBranch(z) {
		if children, ok := Children(z); ok {
			if len(children) == 0 {
				return z, false
			}
			pnodes := []Node{z.focus}
			if z.path != nil {
				pnodes = append(pnodes, z.path.pnodes...)
			}
			return Zipper{
				focus: children[0],
				path: &path{
					right:  children[1:],
					parent: z.path,
					pnodes: pnodes,
				},
			}, true
		}
		return z, false
	} else {
		return z, false
	}
}

// Up moves to the parent of the focused node. If any changes were made to the
// current node or its descendants, Up reconstructs the parent by calling
// [Branch.WithChildren] with the updated child list.
// It returns z, false when z is at the root.
func Up(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}

	focus := z.path.pnodes[0]
	if z.path.changed {
		children := append(append(z.path.left, z.focus), z.path.right...)
		focus = focus.(Branch).WithChildren(children)
		if z.path.parent != nil {
			z.path.parent.changed = true
		}
	}
	return Zipper{
		focus: focus,
		path:  z.path.parent,
	}, true
}

// Right moves to the right sibling of the focused node.
// It returns z, false if z is at the rightmost sibling or at the root.
func Right(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}
	if len(z.path.right) == 0 {
		return z, false
	}
	focus := z.path.right[0]
	left := append(z.path.left, z.focus)
	path := z.path
	path.left = left
	path.right = z.path.right[1:]
	return Zipper{
		focus: focus,
		path:  path,
	}, true
}

// Left moves to the left sibling of the focused node.
// It returns z, false if z is at the leftmost sibling or at the root.
func Left(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}
	if len(z.path.left) == 0 {
		return z, false
	}
	focus := z.path.left[len(z.path.left)-1]
	right := append([]Node{z.focus}, z.path.right...)
	path := z.path
	path.right = right
	path.left = z.path.left[:len(z.path.left)-1]
	return Zipper{
		focus: focus,
		path:  path,
	}, true
}

// Rightmost moves to the rightmost sibling of the focused node. If z is
// already at the rightmost sibling, it returns z, true.
// It returns z, false when z is at the root.
func Rightmost(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}
	if len(z.path.right) == 0 {
		return z, true
	}
	focus := z.path.right[len(z.path.right)-1]
	left := append(append(z.path.left, z.focus), z.path.right[:len(z.path.right)-1]...)
	path := z.path
	path.left = left
	path.right = []Node{}
	return Zipper{
		focus: focus,
		path:  path,
	}, true
}

// Leftmost moves to the leftmost sibling of the focused node. If z is already
// at the leftmost sibling, it returns z, true.
// It returns z, false when z is at the root.
func Leftmost(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}
	if len(z.path.left) == 0 {
		return z, true
	}
	focus := z.path.left[0]
	right := append(append(z.path.left[1:], z.focus), z.path.right...)
	path := z.path
	path.right = right
	path.left = []Node{}
	return Zipper{
		focus: focus,
		path:  path,
	}, true
}

// Root moves to the root of the tree, applying all pending changes along the
// way, and returns a Zipper focused on the root node. When called on an end
// sentinel produced by [Next], Root returns a fresh root Zipper whose focus is
// the fully-accumulated root node (with all edits applied).
//
// Root always returns true.
func Root(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, true
	}
	z, _ = Up(z)
	return Root(z)
}
