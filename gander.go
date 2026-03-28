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

func IsEnd(z Zipper) bool {
	return z.end
}
