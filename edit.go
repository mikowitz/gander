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
