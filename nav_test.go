// ABOUTME: Tests for Down and Up navigation on the gander Zipper.
// ABOUTME: Verifies focus movement, round-tripping, and the WithChildren optimization via CountingBranch.

package gander_test

import (
	"log"
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
				z := gander.NewZipper(tc.node)
				_, ok := gander.Down(z)
				assert.False(t, ok)
			})
		}
	})

	t.Run("focuses the leftmost child", func(t *testing.T) {
		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b, c}})

		child, ok := gander.Down(z)
		require.True(t, ok)

		focused, ok := gander.Focus(child).(StringLeaf)
		require.True(t, ok)
		assert.True(t, focused.Equal(a))
	})

	t.Run("right siblings are preserved on Up after Down", func(t *testing.T) {
		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		z, ok := gander.Down(z)
		require.True(t, ok)

		z, ok = gander.Up(z)
		require.True(t, ok)

		children, ok := gander.Children(z)
		require.True(t, ok)
		require.Len(t, children, 3)
		assert.True(t, children[0].(StringLeaf).Equal(a))
		assert.True(t, children[1].(StringLeaf).Equal(b))
		assert.True(t, children[2].(StringLeaf).Equal(c))
	})
}

func TestUp(t *testing.T) {
	t.Run("returns false at root", func(t *testing.T) {
		z := gander.NewZipper(StringLeaf{Value: "root"})
		_, ok := gander.Up(z)
		assert.False(t, ok)
	})

	t.Run("returns to parent after Down", func(t *testing.T) {
		a := StringLeaf{Value: "a"}
		root := ListBranch{Items: []gander.Node{a, StringLeaf{Value: "b"}}}
		z := gander.NewZipper(root)

		z, ok := gander.Down(z)
		require.True(t, ok)

		z, ok = gander.Up(z)
		require.True(t, ok)

		focused, ok := gander.Focus(z).(ListBranch)
		require.True(t, ok)
		assert.True(t, focused.Equal(root))
	})

	t.Run("does not call WithChildren when no edits were made", func(t *testing.T) {
		cb := NewCountingBranch([]gander.Node{StringLeaf{Value: "a"}, StringLeaf{Value: "b"}})
		z := gander.NewZipper(cb)

		z, ok := gander.Down(z)
		require.True(t, ok)

		_, ok = gander.Up(z)
		require.True(t, ok)

		assert.Equal(t, 0, cb.MakeCount(), "WithChildren must not be called on an unmodified Up")
	})

	t.Run("nested Down → Down → Up → Up returns to root", func(t *testing.T) {
		x := StringLeaf{Value: "x"}
		y := StringLeaf{Value: "y"}
		a := ListBranch{Items: []gander.Node{x, y}}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}

		z := gander.NewZipper(root)

		z, ok := gander.Down(z) // focus: a
		require.True(t, ok)

		z, ok = gander.Down(z) // focus: x
		require.True(t, ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		require.True(t, ok)
		assert.True(t, focused.Equal(x))

		z, ok = gander.Up(z) // focus: a
		require.True(t, ok)

		focused2, ok := gander.Focus(z).(ListBranch)
		require.True(t, ok)
		assert.True(t, focused2.Equal(a))

		z, ok = gander.Up(z) // focus: root
		require.True(t, ok)

		focused3, ok := gander.Focus(z).(ListBranch)
		require.True(t, ok)
		assert.True(t, focused3.Equal(root))
	})
}

func TestRight(t *testing.T) {
	t.Run("Down then Right focuses the second child", func(t *testing.T) {
		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b, c}})

		z, ok := gander.Down(z)
		require.True(t, ok)

		z, ok = gander.Right(z)
		require.True(t, ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		require.True(t, ok)
		assert.True(t, focused.Equal(b))
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
				z := tc.setup()
				_, ok := gander.Right(z)
				assert.False(t, ok)
			})
		}
	})
}

func TestLeft(t *testing.T) {
	t.Run("Right then Left round-trips to the same node", func(t *testing.T) {
		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b}})

		z, ok := gander.Down(z)
		require.True(t, ok)

		z, ok = gander.Right(z)
		require.True(t, ok)

		z, ok = gander.Left(z)
		require.True(t, ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		require.True(t, ok)
		assert.True(t, focused.Equal(a))
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
				z := tc.setup()
				_, ok := gander.Left(z)
				assert.False(t, ok)
			})
		}
	})
}

func TestLeftmost(t *testing.T) {
	t.Run("moves to the first sibling from any position", func(t *testing.T) {
		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b, c}})

		// Navigate to the rightmost child (c) and call Leftmost.
		z, ok := gander.Down(z)
		require.True(t, ok)
		z, ok = gander.Right(z)
		require.True(t, ok)
		z, ok = gander.Right(z)
		require.True(t, ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		require.True(t, ok)
		assert.True(t, focused.Equal(c), "setup: should be focused on c")

		z, ok = gander.Leftmost(z)
		require.True(t, ok)

		focused, ok = gander.Focus(z).(StringLeaf)
		require.True(t, ok)
		assert.True(t, focused.Equal(a))
	})

	t.Run("returns self when already at leftmost", func(t *testing.T) {
		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b}})

		z, ok := gander.Down(z) // focused on a (leftmost)
		require.True(t, ok)

		z, ok = gander.Leftmost(z)
		require.True(t, ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		require.True(t, ok)
		assert.True(t, focused.Equal(a))
	})

	t.Run("returns false at root", func(t *testing.T) {
		z := gander.NewZipper(ListBranch{Items: []gander.Node{StringLeaf{Value: "a"}}})
		_, ok := gander.Leftmost(z)
		assert.False(t, ok)
	})
}

func TestRightmost(t *testing.T) {
	t.Run("moves to the last sibling from any position", func(t *testing.T) {
		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b, c}})

		// Navigate to the leftmost child (a) and call Rightmost.
		z, ok := gander.Down(z)
		require.True(t, ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		require.True(t, ok)
		assert.True(t, focused.Equal(a), "setup: should be focused on a")

		z, ok = gander.Rightmost(z)
		require.True(t, ok)

		focused, ok = gander.Focus(z).(StringLeaf)
		log.Println(focused, ok)
		require.True(t, ok)
		assert.True(t, focused.Equal(c))
	})

	t.Run("returns self when already at rightmost", func(t *testing.T) {
		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		z := gander.NewZipper(ListBranch{Items: []gander.Node{a, b}})

		z, ok := gander.Down(z)
		require.True(t, ok)
		z, ok = gander.Right(z) // focused on b (rightmost)
		require.True(t, ok)

		z, ok = gander.Rightmost(z)
		require.True(t, ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		require.True(t, ok)
		assert.True(t, focused.Equal(b))
	})

	t.Run("returns false at root", func(t *testing.T) {
		z := gander.NewZipper(ListBranch{Items: []gander.Node{StringLeaf{Value: "a"}}})
		_, ok := gander.Rightmost(z)
		assert.False(t, ok)
	})
}
