// ABOUTME: Tests for the Remove editing operation.
// ABOUTME: Verifies removal semantics, depth-first predecessor navigation, and tree reconstruction after removal.

package gander_test

import (
	"fmt"
	"testing"

	"github.com/mikowitz/gander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemoveAtRoot verifies that Remove returns false when called at root,
// since there is no predecessor for the root node.
func TestRemoveAtRoot(t *testing.T) {
	t.Run("returns false when called at root", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		root := ListBranch{Items: []gander.Node{
			StringLeaf{Value: "a"},
			StringLeaf{Value: "b"},
		}}
		z := gander.NewZipper(root)

		_, ok := gander.Remove(z)
		req.False(ok)

		// Focus must be unchanged.
		focused, ok := gander.Focus(z).(ListBranch)
		req.True(ok)
		asrt.True(focused.Equal(root))
	})
}

// TestRemoveMovesToDepthFirstPredecessor verifies that when the focused node has left siblings,
// Remove moves focus to the rightmost descendant of the nearest left sibling.
//
// Tree structure used:
//
//	root → [left_branch → [x, y → [p, q]], current]
//
// Focus is on "current". The nearest left sibling is "left_branch", whose rightmost descendant
// is "q" (right child of "y", which is right child of "left_branch"). After Remove, focus
// should be on "q".
func TestRemoveMovesToDepthFirstPredecessor(t *testing.T) {
	t.Run("moves to rightmost descendant of left sibling when left siblings exist", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		// Build: root → [left_branch → [x, y → [p, q]], current]
		p := StringLeaf{Value: "p"}
		q := StringLeaf{Value: "q"}
		x := StringLeaf{Value: "x"}
		y := ListBranch{Items: []gander.Node{p, q}}
		leftBranch := ListBranch{Items: []gander.Node{x, y}}
		current := StringLeaf{Value: "current"}
		root := ListBranch{Items: []gander.Node{leftBranch, current}}

		z := gander.NewZipper(root)

		// Navigate to "current": Down to leftBranch, Right to current.
		z, ok := gander.Down(z) // focus: leftBranch
		req.True(ok)
		z, ok = gander.Right(z) // focus: current
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(current), "precondition: focus is on current before Remove")

		// Remove "current"; predecessor is the rightmost descendant of leftBranch,
		// which is "q" (rightmost child of "y", which is rightmost child of leftBranch).
		z, ok = gander.Remove(z)
		req.True(ok)

		result, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(result.Equal(q), "after Remove, focus should be on q (rightmost descendant of left sibling)")
	})
}

// TestRemoveAtLeftmostChildMovesToParent verifies that when the focused node is the leftmost
// child (no left siblings), Remove moves focus to the parent, which is reconstructed with
// only the right siblings as its new children.
func TestRemoveAtLeftmostChildMovesToParent(t *testing.T) {
	t.Run("moves to parent with remaining children when leftmost child is removed", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		// Navigate to the leftmost child "a".
		z, ok := gander.Down(z) // focus: "a"
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		req.True(focused.Equal(a), "precondition: focus is on a")

		// Remove "a"; since it is leftmost, focus moves to parent with [b, c] as children.
		z, ok = gander.Remove(z)
		req.True(ok)
		fmt.Println(z)

		// Focus should now be on the parent (a ListBranch).
		parent, ok := gander.Focus(z).(ListBranch)
		req.True(ok)

		fmt.Println(parent)

		expected := ListBranch{Items: []gander.Node{b, c}}
		asrt.True(parent.Equal(expected), "parent should now contain only b and c")
	})
}

// TestRemoveOnlyChild verifies that removing the only child of a branch leaves
// the parent with an empty children slice.
func TestRemoveOnlyChild(t *testing.T) {
	t.Run("leaves parent with empty children when only child is removed", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		onlyChild := StringLeaf{Value: "only"}
		root := ListBranch{Items: []gander.Node{onlyChild}}
		z := gander.NewZipper(root)

		// Navigate to the only child.
		z, ok := gander.Down(z) // focus: onlyChild
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(onlyChild), "precondition: focus is on the only child")

		// Remove the only child; focus moves to parent with empty children.
		z, ok = gander.Remove(z)
		req.True(ok)

		// Focus should be the parent with no children.
		parent, ok := gander.Focus(z).(ListBranch)
		req.True(ok)

		expected := ListBranch{Items: []gander.Node{}}
		asrt.True(parent.Equal(expected), "parent should now have empty children")

		children, ok := gander.Children(z)
		req.True(ok)
		asrt.Empty(children, "children of parent should be empty after removing the only child")
	})
}

// TestRemoveThenRoot verifies that after a Remove, calling Root reflects the removal
// in the final tree — the removed node is absent from the reconstructed structure.
func TestRemoveThenRoot(t *testing.T) {
	t.Run("Root after Remove shows the node is gone from the tree", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		// Navigate to "b" (second child) and remove it.
		z, ok := gander.Down(z) // focus: "a"
		req.True(ok)
		z, ok = gander.Right(z) // focus: "b"
		req.True(ok)

		z, ok = gander.Remove(z)
		req.True(ok)

		// Zip back to root and verify "b" is gone.
		rootZ, ok := gander.Root(z)
		req.True(ok)

		children, ok := gander.Children(rootZ)
		req.True(ok)
		req.Len(children, 2, "root should have 2 children after removing b")
		asrt.True(children[0].(StringLeaf).Equal(a), "first child should be a")
		asrt.True(children[1].(StringLeaf).Equal(c), "second child should be c")
	})
}
