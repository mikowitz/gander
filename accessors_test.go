// ABOUTME: Tests for the Path, Lefts, and Rights accessor functions.
// ABOUTME: Verifies that ancestor paths and sibling slices are returned in correct tree order.

package gander_test

import (
	"testing"

	"github.com/mikowitz/gander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	req := require.New(t)

	a := StringLeaf{Value: "a"}
	b := StringLeaf{Value: "b"}
	c := StringLeaf{Value: "c"}
	innerBranch := ListBranch{Items: []gander.Node{b, c}}
	root := ListBranch{Items: []gander.Node{innerBranch, a}}

	tests := []struct {
		name     string
		navigate func(z gander.Zipper) gander.Zipper
		wantPath []gander.Node
	}{
		{
			name:     "at root returns empty slice",
			navigate: func(z gander.Zipper) gander.Zipper { return z },
			wantPath: []gander.Node{},
		},
		{
			name: "after Down returns slice containing root",
			navigate: func(z gander.Zipper) gander.Zipper {
				z, ok := gander.Down(z)
				req.True(ok)
				return z
			},
			wantPath: []gander.Node{root},
		},
		{
			name: "after Down then Down returns slice containing root and first child",
			navigate: func(z gander.Zipper) gander.Zipper {
				z, ok := gander.Down(z)
				req.True(ok)
				z, ok = gander.Down(z)
				req.True(ok)
				return z
			},
			wantPath: []gander.Node{root, innerBranch},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			asrt := assert.New(t)
			req := require.New(t)

			z := gander.NewZipper(root)
			z = tc.navigate(z)
			got := gander.Path(z)
			req.Len(got, len(tc.wantPath))
			for i, want := range tc.wantPath {
				asrt.True(want.(interface {
					Equal(gander.Node) bool
				}).Equal(got[i]),
					"path[%d]: expected %v, got %v", i, want, got[i],
				)
			}
		})
	}
}

func TestLefts(t *testing.T) {
	a := StringLeaf{Value: "a"}
	b := StringLeaf{Value: "b"}
	c := StringLeaf{Value: "c"}
	root := ListBranch{Items: []gander.Node{a, b, c}}

	t.Run("at root returns nil", func(t *testing.T) {
		asrt := assert.New(t)

		z := gander.NewZipper(root)
		asrt.Nil(gander.Lefts(z))
	})

	t.Run("after Down returns empty slice at leftmost child", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		z := gander.NewZipper(root)
		z, ok := gander.Down(z)
		req.True(ok)

		got := gander.Lefts(z)
		asrt.Empty(got)
	})

	t.Run("after Down then Right returns slice containing first child", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		z := gander.NewZipper(root)
		z, ok := gander.Down(z)
		req.True(ok)
		z, ok = gander.Right(z)
		req.True(ok)

		got := gander.Lefts(z)
		req.Len(got, 1)
		leaf, ok := got[0].(StringLeaf)
		req.True(ok)
		asrt.True(leaf.Equal(a))
	})

	t.Run("returns siblings in left-to-right order after navigating right", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		// [a, b, c, d, e] — navigate right to c; Lefts must be [a, b], not [b, a]
		d := StringLeaf{Value: "d"}
		e := StringLeaf{Value: "e"}
		wide := ListBranch{Items: []gander.Node{a, b, c, d, e}}
		z := gander.NewZipper(wide)
		z, ok := gander.Down(z)
		req.True(ok)
		z, ok = gander.Right(z)
		req.True(ok)
		z, ok = gander.Right(z)
		req.True(ok) // now at c

		got := gander.Lefts(z)
		req.Len(got, 2)
		asrt.True(got[0].(StringLeaf).Equal(a), "Lefts[0] should be a")
		asrt.True(got[1].(StringLeaf).Equal(b), "Lefts[1] should be b")
	})

	t.Run("returns siblings in left-to-right order after navigating left", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		// [a, b, c, d, e] — navigate to e then left to c; Lefts must be [a, b], not [b, a]
		d := StringLeaf{Value: "d"}
		e := StringLeaf{Value: "e"}
		wide := ListBranch{Items: []gander.Node{a, b, c, d, e}}
		z := gander.NewZipper(wide)
		z, ok := gander.Down(z)
		req.True(ok)
		z, ok = gander.Rightmost(z)
		req.True(ok) // now at e
		z, ok = gander.Left(z)
		req.True(ok) // now at d
		z, ok = gander.Left(z)
		req.True(ok) // now at c

		got := gander.Lefts(z)
		req.Len(got, 2)
		asrt.True(got[0].(StringLeaf).Equal(a), "Lefts[0] should be a")
		asrt.True(got[1].(StringLeaf).Equal(b), "Lefts[1] should be b")
	})
}

func TestRights(t *testing.T) {
	a := StringLeaf{Value: "a"}
	b := StringLeaf{Value: "b"}
	c := StringLeaf{Value: "c"}
	root := ListBranch{Items: []gander.Node{a, b, c}}

	t.Run("at root returns nil", func(t *testing.T) {
		asrt := assert.New(t)

		z := gander.NewZipper(root)
		asrt.Nil(gander.Rights(z))
	})

	t.Run("after Down returns remaining right siblings", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		z := gander.NewZipper(root)
		z, ok := gander.Down(z)
		req.True(ok)

		got := gander.Rights(z)
		req.Len(got, 2)

		leafB, ok := got[0].(StringLeaf)
		req.True(ok)
		asrt.True(leafB.Equal(b))

		leafC, ok := got[1].(StringLeaf)
		req.True(ok)
		asrt.True(leafC.Equal(c))
	})

	t.Run("after Rightmost returns empty slice", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		z := gander.NewZipper(root)
		z, ok := gander.Down(z)
		req.True(ok)
		z, ok = gander.Rightmost(z)
		req.True(ok)

		got := gander.Rights(z)
		asrt.Empty(got)
	})

	t.Run("returns siblings in left-to-right order after navigating right", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		// [a, b, c, d, e] — navigate right to c; Rights must be [d, e], not [e, d]
		d := StringLeaf{Value: "d"}
		e := StringLeaf{Value: "e"}
		wide := ListBranch{Items: []gander.Node{a, b, c, d, e}}
		z := gander.NewZipper(wide)
		z, ok := gander.Down(z)
		req.True(ok)
		z, ok = gander.Right(z)
		req.True(ok)
		z, ok = gander.Right(z)
		req.True(ok) // now at c

		got := gander.Rights(z)
		req.Len(got, 2)
		asrt.True(got[0].(StringLeaf).Equal(d), "Rights[0] should be d")
		asrt.True(got[1].(StringLeaf).Equal(e), "Rights[1] should be e")
	})

	t.Run("returns siblings in left-to-right order after navigating left", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		// [a, b, c, d, e] — navigate to e then left to c; Rights must be [d, e], not [e, d]
		d := StringLeaf{Value: "d"}
		e := StringLeaf{Value: "e"}
		wide := ListBranch{Items: []gander.Node{a, b, c, d, e}}
		z := gander.NewZipper(wide)
		z, ok := gander.Down(z)
		req.True(ok)
		z, ok = gander.Rightmost(z)
		req.True(ok) // now at e
		z, ok = gander.Left(z)
		req.True(ok) // now at d
		z, ok = gander.Left(z)
		req.True(ok) // now at c

		got := gander.Rights(z)
		req.Len(got, 2)
		asrt.True(got[0].(StringLeaf).Equal(d), "Rights[0] should be d")
		asrt.True(got[1].(StringLeaf).Equal(e), "Rights[1] should be e")
	})
}
