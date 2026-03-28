package gander

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
