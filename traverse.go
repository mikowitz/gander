// ABOUTME: Implements the depth-first traversal functions Next and Prev.
// ABOUTME: These functions walk a tree in depth-first order and are designed to be used with IsEnd.

package gander

// Next moves to the next location in depth-first order. If the focused node
// is a [Branch] with children, Next descends to the leftmost child. Otherwise
// it moves to the right sibling, or walks up the tree until a right sibling is
// found. When the traversal is exhausted, Next returns an end sentinel; use
// [IsEnd] to detect it. Calling Next on the end sentinel returns the end
// sentinel unchanged.
func Next(z Zipper) Zipper {
	if IsEnd(z) {
		return z
	}
	if d, dOK := Down(z); dOK {
		return d
	}
	if r, rOK := Right(z); rOK {
		return r
	}
	return recurNext(z)
}

func recurNext(z Zipper) Zipper {
	u, uOK := Up(z)
	if uOK {
		r, rOK := Right(u)
		if rOK {
			return r
		}
		return recurNext(u)
	}

	return Zipper{
		focus: z.focus,
		path:  z.path,
		end:   true,
	}
}

// Prev moves to the previous location in depth-first order, reversing the
// traversal order of [Next]. If the focused node has left siblings, Prev moves
// to the rightmost descendant of the nearest left sibling. Otherwise it moves
// up to the parent.
// It returns z, false when z is at the root or when z is the end sentinel.
func Prev(z Zipper) (Zipper, bool) {
	if z.path == nil {
		return z, false
	}
	if IsEnd(z) {
		return z, false
	}
	if l, lOK := Left(z); lOK {
		return recurPrev(l)
	}
	if u, uOK := Up(z); uOK {
		return u, true
	}
	return z, false
}

func recurPrev(z Zipper) (Zipper, bool) {
	if d, dOK := Down(z); dOK {
		n, _ := Rightmost(d)
		return recurPrev(n)
	}
	return z, true
}
