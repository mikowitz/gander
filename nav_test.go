// ABOUTME: Tests for Down and Up navigation on the gander Zipper.
// ABOUTME: Verifies focus movement, round-tripping, and the WithChildren optimization via CountingBranch.

package gander_test

import (
	"testing"

	"github.com/mikowitz/gander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDown(t *testing.T) {
	t.Run("returns false for non-navigable nodes", func(t *testing.T) {
		tests := []struct {
			name string
			node gander.Node
		}{
			{
				name: "leaf node",
				node: StringLeaf{Value: "x"},
			},
			{
				name: "empty branch",
				node: ListBranch{Items: []gander.Node{}},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				asrt := assert.New(t)

				z := gander.NewZipper(tc.node)
				_, ok := gander.Down(z)
				asrt.False(ok)
			})
		}
	})

	t.Run("focuses the leftmost child", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b, c}})

		child, ok := gander.Down(z)
		req.True(ok)

		focused, ok := gander.Focus(child).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a))
	})

	t.Run("right siblings are preserved on Up after Down", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		z, ok := gander.Down(z)
		req.True(ok)

		z, ok = gander.Up(z)
		req.True(ok)

		children, ok := gander.Children(z)
		req.True(ok)
		req.Len(children, 3)
		asrt.True(children[0].(StringLeaf).Equal(a))
		asrt.True(children[1].(StringLeaf).Equal(b))
		asrt.True(children[2].(StringLeaf).Equal(c))
	})
}

func TestUp(t *testing.T) {
	t.Run("returns false at root", func(t *testing.T) {
		asrt := assert.New(t)

		z := gander.NewZipper(StringLeaf{Value: "root"})
		_, ok := gander.Up(z)
		asrt.False(ok)
	})

	t.Run("returns to parent after Down", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		root := ListBranch{Items: []gander.Node{a, StringLeaf{Value: "b"}}}
		z := gander.NewZipper(root)

		z, ok := gander.Down(z)
		req.True(ok)

		z, ok = gander.Up(z)
		req.True(ok)

		focused, ok := gander.Focus(z).(ListBranch)
		req.True(ok)
		asrt.True(focused.Equal(root))
	})

	t.Run("does not call WithChildren when no edits were made", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		cb := NewCountingBranch([]gander.Node{StringLeaf{Value: "a"}, StringLeaf{Value: "b"}})
		z := gander.NewZipper(cb)

		z, ok := gander.Down(z)
		req.True(ok)

		_, ok = gander.Up(z)
		req.True(ok)

		asrt.Equal(0, cb.MakeCount(), "WithChildren must not be called on an unmodified Up")
	})

	t.Run("nested Down → Down → Up → Up returns to root", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		x := StringLeaf{Value: "x"}
		y := StringLeaf{Value: "y"}
		a := ListBranch{Items: []gander.Node{x, y}}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}

		z := gander.NewZipper(root)

		z, ok := gander.Down(z) // focus: a
		req.True(ok)

		z, ok = gander.Down(z) // focus: x
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(x))

		z, ok = gander.Up(z) // focus: a
		req.True(ok)

		focused2, ok := gander.Focus(z).(ListBranch)
		req.True(ok)
		asrt.True(focused2.Equal(a))

		z, ok = gander.Up(z) // focus: root
		req.True(ok)

		focused3, ok := gander.Focus(z).(ListBranch)
		req.True(ok)
		asrt.True(focused3.Equal(root))
	})
}

func TestRight(t *testing.T) {
	t.Run("Down then Right focuses the second child", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b, c}})

		z, ok := gander.Down(z)
		req.True(ok)

		z, ok = gander.Right(z)
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(b))
	})

	t.Run("returns false when navigation is not possible", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func() gander.Zipper
		}{
			{
				name: "at rightmost sibling",
				setup: func() gander.Zipper {
					a := StringLeaf{Value: "a"}
					b := StringLeaf{Value: "b"}
					z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b}})
					z, _ = gander.Down(z)
					z, _ = gander.Right(z) // now at b, the rightmost
					return z
				},
			},
			{
				name: "at root",
				setup: func() gander.Zipper {
					return gander.NewZipper(ListBranch{Items: []gander.Node{StringLeaf{Value: "a"}}})
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				asrt := assert.New(t)

				z := tc.setup()
				_, ok := gander.Right(z)
				asrt.False(ok)
			})
		}
	})
}

func TestLeft(t *testing.T) {
	t.Run("Right then Left round-trips to the same node", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b}})

		z, ok := gander.Down(z)
		req.True(ok)

		z, ok = gander.Right(z)
		req.True(ok)

		z, ok = gander.Left(z)
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a))
	})

	t.Run("returns false when navigation is not possible", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func() gander.Zipper
		}{
			{
				name: "at leftmost sibling",
				setup: func() gander.Zipper {
					a := StringLeaf{Value: "a"}
					b := StringLeaf{Value: "b"}
					z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b}})
					z, _ = gander.Down(z) // now at a, the leftmost
					return z
				},
			},
			{
				name: "at root",
				setup: func() gander.Zipper {
					return gander.NewZipper(ListBranch{Items: []gander.Node{StringLeaf{Value: "a"}}})
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				asrt := assert.New(t)

				z := tc.setup()
				_, ok := gander.Left(z)
				asrt.False(ok)
			})
		}
	})
}

func TestLeftmost(t *testing.T) {
	t.Run("moves to the first sibling from any position", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b, c}})

		// Navigate to the rightmost child (c) and call Leftmost.
		z, ok := gander.Down(z)
		req.True(ok)
		z, ok = gander.Right(z)
		req.True(ok)
		z, ok = gander.Right(z)
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(c), "setup: should be focused on c")

		z, ok = gander.Leftmost(z)
		req.True(ok)

		focused, ok = gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a))
	})

	t.Run("returns self when already at leftmost", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b}})

		z, ok := gander.Down(z) // focused on a (leftmost)
		req.True(ok)

		z, ok = gander.Leftmost(z)
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a))
	})

	t.Run("returns false at root", func(t *testing.T) {
		asrt := assert.New(t)

		z := gander.NewZipper(ListBranch{Items: []gander.Node{StringLeaf{Value: "a"}}})
		_, ok := gander.Leftmost(z)
		asrt.False(ok)
	})
}

func TestRightmost(t *testing.T) {
	t.Run("moves to the last sibling from any position", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b, c}})

		// Navigate to the leftmost child (a) and call Rightmost.
		z, ok := gander.Down(z)
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a), "setup: should be focused on a")

		z, ok = gander.Rightmost(z)
		req.True(ok)

		focused, ok = gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(c))
	})

	t.Run("returns self when already at rightmost", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b}})

		z, ok := gander.Down(z)
		req.True(ok)
		z, ok = gander.Right(z) // focused on b (rightmost)
		req.True(ok)

		z, ok = gander.Rightmost(z)
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(b))
	})

	t.Run("returns false at root", func(t *testing.T) {
		asrt := assert.New(t)

		z := gander.NewZipper(ListBranch{Items: []gander.Node{StringLeaf{Value: "a"}}})
		_, ok := gander.Rightmost(z)
		asrt.False(ok)
	})
}
