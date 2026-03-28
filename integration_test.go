// ABOUTME: Integration tests verifying the full gander Zipper contract.
// ABOUTME: Covers immutability, round-trips, structural sharing, complex edits, and heterogeneous trees.

package gander_test

import (
	"reflect"
	"testing"

	"github.com/mikowitz/gander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntLeaf is a leaf node holding an integer value, used in heterogeneous tree tests.
type IntLeaf struct {
	gander.BaseNode
	Value int
}

func (i IntLeaf) Equal(other gander.Node) bool {
	o, ok := other.(IntLeaf)
	return ok && i.Value == o.Value
}

func TestImmutability(t *testing.T) {
	t.Run("held zipper is unaffected by subsequent navigation and edits", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		snapshot, ok := gander.Down(z)
		req.True(ok, "Down should succeed")

		// Navigate Right → Right from snapshot to reach c, then Replace with "C"
		edited := snapshot
		edited, ok = gander.Right(edited)
		req.True(ok)
		edited, ok = gander.Right(edited)
		req.True(ok)
		edited, ok = gander.Replace(edited, StringLeaf{Value: "C"})
		req.True(ok)
		_, ok = gander.Root(edited)
		req.True(ok)

		// The held snapshot should be focused on a, unchanged
		focused := gander.Focus(snapshot).(StringLeaf)
		asrt.True(focused.Equal(a), "held snapshot focus should remain on a after navigation and edits on derived zipper")
	})
}

func TestRoundTrip(t *testing.T) {
	t.Run("full Next traversal without editing returns original root", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		x := StringLeaf{Value: "x"}
		y := StringLeaf{Value: "y"}
		bigA := ListBranch{Items: []gander.Node{x, y}}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{bigA, b}}
		z := gander.NewZipper(root)

		for !gander.IsEnd(z) {
			z = gander.Next(z)
		}

		rootZ, ok := gander.Root(z)
		req.True(ok, "Root on end sentinel should succeed")

		asrt.True(reflect.DeepEqual(gander.Focus(rootZ), root), "root after read-only traversal should equal original root")
	})

	t.Run("full Next traversal does not call WithChildren", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := NewCountingBranch([]gander.Node{a, b})
		z := gander.NewZipper(root)

		for !gander.IsEnd(z) {
			z = gander.Next(z)
		}

		_, ok := gander.Root(z)
		req.True(ok, "Root on end sentinel should succeed")

		asrt.Equal(0, root.MakeCount(), "WithChildren should not be called during a read-only traversal")
	})
}

func TestStructuralSharing(t *testing.T) {
	t.Run("editing at deepest level calls WithChildren exactly once per ancestor", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		b := NewCountingBranch([]gander.Node{StringLeaf{Value: "leaf"}})
		a := NewCountingBranch([]gander.Node{b})
		root := NewCountingBranch([]gander.Node{a})
		z := gander.NewZipper(root)

		z, ok := gander.Down(z)
		req.True(ok, "first Down should succeed")
		z, ok = gander.Down(z)
		req.True(ok, "second Down should succeed")
		z, ok = gander.Down(z)
		req.True(ok, "third Down should succeed")

		z, ok = gander.Replace(z, StringLeaf{Value: "changed"})
		req.True(ok, "Replace should succeed")

		_, ok = gander.Root(z)
		req.True(ok, "Root should succeed")

		asrt.Equal(1, b.MakeCount(), "b.WithChildren should be called exactly once")
		asrt.Equal(1, a.MakeCount(), "a.WithChildren should be called exactly once")
		asrt.Equal(1, root.MakeCount(), "root.WithChildren should be called exactly once")
	})
}

func TestComplexEdit(t *testing.T) {
	t.Run("replaces all * with / in nested expression tree", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		inner1 := ListBranch{Items: []gander.Node{
			StringLeaf{Value: "a"}, StringLeaf{Value: "*"}, StringLeaf{Value: "b"},
		}}
		inner2 := ListBranch{Items: []gander.Node{
			StringLeaf{Value: "c"}, StringLeaf{Value: "*"}, StringLeaf{Value: "d"},
		}}
		root := ListBranch{Items: []gander.Node{inner1, StringLeaf{Value: "+"}, inner2}}
		z := gander.NewZipper(root)

		for !gander.IsEnd(z) {
			if leaf, ok := gander.Focus(z).(StringLeaf); ok && leaf.Value == "*" {
				z, _ = gander.Replace(z, StringLeaf{Value: "/"})
			}
			z = gander.Next(z)
		}

		rootZ, ok := gander.Root(z)
		req.True(ok, "Root on end sentinel should succeed")

		expected := ListBranch{Items: []gander.Node{
			ListBranch{Items: []gander.Node{
				StringLeaf{Value: "a"}, StringLeaf{Value: "/"}, StringLeaf{Value: "b"},
			}},
			StringLeaf{Value: "+"},
			ListBranch{Items: []gander.Node{
				StringLeaf{Value: "c"}, StringLeaf{Value: "/"}, StringLeaf{Value: "d"},
			}},
		}}
		focusedRoot, ok := gander.Focus(rootZ).(ListBranch)
		req.True(ok, "focused node after Root should be a ListBranch")
		asrt.True(focusedRoot.Equal(expected), "all * nodes should be replaced with /")
	})
}

func TestRemoveDuringTraversal(t *testing.T) {
	t.Run("removing a node during Next walk produces correct Root", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		for !gander.IsEnd(z) {
			if leaf, ok := gander.Focus(z).(StringLeaf); ok && leaf.Value == "b" {
				z, _ = gander.Remove(z)
				// After Remove, focus is on a (depth-first predecessor)
			}
			z = gander.Next(z)
		}

		rootZ, ok := gander.Root(z)
		req.True(ok, "Root on end sentinel should succeed")

		children, ok := gander.Children(rootZ)
		req.True(ok, "root should be a branch")
		req.Len(children, 2, "root should have exactly 2 children after removing b")
		asrt.True(children[0].(StringLeaf).Equal(a), "first child should be a")
		asrt.True(children[1].(StringLeaf).Equal(c), "second child should be c")
	})
}

func TestHeterogeneousTree(t *testing.T) {
	t.Run("navigating and editing a tree with mixed node types works correctly", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		inner := ListBranch{Items: []gander.Node{IntLeaf{Value: 2}, StringLeaf{Value: "y"}}}
		root := ListBranch{Items: []gander.Node{
			IntLeaf{Value: 1},
			StringLeaf{Value: "x"},
			inner,
		}}
		z := gander.NewZipper(root)

		z, ok := gander.Down(z)
		req.True(ok, "Down should succeed")
		z, ok = gander.Right(z)
		req.True(ok, "Right should succeed")

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok, "focus should be a StringLeaf")
		req.True(focused.Equal(StringLeaf{Value: "x"}), "focus should be x")

		z, ok = gander.Replace(z, StringLeaf{Value: "X"})
		req.True(ok, "Replace should succeed")

		rootZ, ok := gander.Root(z)
		req.True(ok, "Root should succeed")

		children, ok := gander.Children(rootZ)
		req.True(ok, "root should be a branch")
		req.Len(children, 3, "root should have 3 children")
		asrt.True(children[0].(IntLeaf).Equal(IntLeaf{Value: 1}), "first child should be IntLeaf{1}")
		asrt.True(children[1].(StringLeaf).Equal(StringLeaf{Value: "X"}), "second child should be StringLeaf{X}")
		asrt.True(children[2].(ListBranch).Equal(inner), "third child should be the inner branch unchanged")
	})
}

func TestEmptyBranchHandling(t *testing.T) {
	t.Run("navigating around an empty branch preserves structure", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		emptyBranch := ListBranch{Items: []gander.Node{}}
		leaf := StringLeaf{Value: "leaf"}
		root := ListBranch{Items: []gander.Node{emptyBranch, leaf}}
		z := gander.NewZipper(root)

		z, ok := gander.Down(z)
		req.True(ok, "Down should succeed")

		focused, ok := gander.Focus(z).(ListBranch)
		req.True(ok, "focus should be a ListBranch")
		asrt.True(focused.Equal(emptyBranch), "focus should be emptyBranch")

		// Attempting to go Down into an empty branch should fail
		_, ok = gander.Down(z)
		asrt.False(ok, "Down into empty branch should return false")

		z, ok = gander.Right(z)
		req.True(ok, "Right should succeed")

		focused2, ok := gander.Focus(z).(StringLeaf)
		req.True(ok, "focus should be a StringLeaf")
		asrt.True(focused2.Equal(leaf), "focus should be leaf")

		z, ok = gander.Up(z)
		req.True(ok, "Up should succeed")

		children, ok := gander.Children(z)
		req.True(ok, "root should be a branch")
		req.Len(children, 2, "root should have 2 children")
		asrt.True(children[0].(ListBranch).Equal(emptyBranch), "first child should be emptyBranch")
		asrt.True(children[1].(StringLeaf).Equal(leaf), "second child should be leaf")
	})
}
