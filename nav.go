package gander

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

func Root(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, true
	}
	if z, ok := Up(z); ok {
		return Root(z)
	}
	return z, false
}
