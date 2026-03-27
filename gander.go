package gander

type Node interface {
	node()
}

type Leaf interface {
	Node
}

type Branch interface {
	Node
	Children() []Node
	WithChildren(children []Node) Branch
}

type BaseNode struct{}

func (BaseNode) node() {}

type path struct {
	left, right []Node
	parent      *path
	pnodes      []Node
	changed     bool
}

type Zipper struct {
	focus Node
	path  *path
	end   bool
}

func NewZipper(root Node) Zipper {
	return Zipper{focus: root}
}

func Focus(z Zipper) Node {
	return z.focus
}

func IsBranch(z Zipper) bool {
	switch z.focus.(type) {
	case Branch:
		return true
	default:
		return false
	}
}

func Children(z Zipper) ([]Node, bool) {
	if IsBranch(z) {
		return z.focus.(Branch).Children(), true
	} else {
		return nil, false
	}
}

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
	z.path.left = left
	z.path.right = z.path.right[1:]
	return Zipper{
		focus: focus,
		path:  z.path,
	}, true
}

func Left(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}
	if len(z.path.left) == 0 {
		return z, false
	}
	focus := z.path.left[0]
	right := append(z.path.right, z.focus)
	z.path.right = right
	z.path.left = z.path.left[1:]
	return Zipper{
		focus: focus,
		path:  z.path,
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
	z.path.left = left
	z.path.right = []Node{}
	return Zipper{
		focus: focus,
		path:  z.path,
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
	z.path.right = right
	z.path.left = []Node{}
	return Zipper{
		focus: focus,
		path:  z.path,
	}, true
}
