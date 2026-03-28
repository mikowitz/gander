// ABOUTME: Implements the Remove function for deleting the focused node from the tree.
// ABOUTME: After removal, the Zipper is repositioned at the depth-first predecessor.

package gander

// Remove removes the focused node and returns a Zipper positioned at the
// depth-first predecessor of the removed node. If the focused node has left
// siblings, the predecessor is the rightmost descendant of the nearest left
// sibling. If the focused node is the leftmost child, the predecessor is the
// parent.
//
// It returns z, false when z is at the root.
func Remove(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}

	if len(z.path.left) > 0 {
		pred := Zipper{
			focus: z.path.left[len(z.path.left)-1],
			path: &path{
				left:    z.path.left[:len(z.path.left)-1],
				right:   z.path.right,
				parent:  z.path.parent,
				pnodes:  z.path.pnodes,
				changed: true,
			},
		}
		pred = findPred(pred)
		return pred, true
	}

	newParent := z.path.pnodes[len(z.path.pnodes)-1]
	newParent = newParent.(Branch).WithChildren(z.path.right)
	parentPath := z.path.parent
	if parentPath != nil {
		parentPath.changed = true
	}
	return Zipper{
		focus: newParent,
		path:  parentPath,
	}, true
}

func findPred(z Zipper) Zipper {
	if IsBranch(z) && len(z.focus.(Branch).Children()) > 0 {
		z, _ := Down(z)
		z, _ = Rightmost(z)
		return findPred(z)
	}
	return z
}
