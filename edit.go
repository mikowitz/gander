package gander

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

func Edit(z Zipper, f func(Node) Node) (Zipper, bool) {
	focus := f(z.focus)
	return Replace(z, focus)
}

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
