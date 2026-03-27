package gander

import "slices"

func Path(z Zipper) []Node {
	if z.path == nil {
		return []Node{}
	}
	ret := make([]Node, len(z.path.pnodes))
	copy(ret, z.path.pnodes)
	slices.Reverse(ret)
	return ret
}

func Lefts(z Zipper) []Node {
	return z.path.left
}

func Rights(z Zipper) []Node {
	return z.path.right
}
