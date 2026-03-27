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
