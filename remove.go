package gander

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
