// ABOUTME: Tests for depth-first traversal operations Next, Prev, and IsEnd.
// ABOUTME: Verifies traversal order, end sentinel behavior, and edit-during-traversal patterns.

package gander_test

import (
	"log"
	"testing"

	"github.com/mikowitz/gander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNext(t *testing.T) {
	t.Run("from root goes to first child", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		z = gander.Next(z)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok, "focus after Next from root should be a StringLeaf")
		asrt.True(focused.Equal(a), "focus after Next from root should be the first child")
	})

	t.Run("visits all nodes in depth-first order", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		// Build tree: root → [A → [x, y], b]
		x := StringLeaf{Value: "x"}
		y := StringLeaf{Value: "y"}
		bigA := ListBranch{Items: []gander.Node{x, y}}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{bigA, b}}
		z := gander.NewZipper(root)

		var visited []string
		for !gander.IsEnd(z) {
			log.Println(gander.Focus(z))
			switch n := gander.Focus(z).(type) {
			case ListBranch:
				if n.Equal(root) {
					visited = append(visited, "root")
				} else {
					visited = append(visited, "A")
				}
			case StringLeaf:
				visited = append(visited, n.Value)
			}
			z = gander.Next(z)
		}

		req.Len(visited, 5, "should visit exactly 5 nodes")
		asrt.Equal([]string{"root", "A", "x", "y", "b"}, visited, "nodes should be visited in depth-first order")
	})

	t.Run("after last node returns end sentinel", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		root := ListBranch{Items: []gander.Node{a}}
		z := gander.NewZipper(root)

		// Navigate to "a" via Down, then call Next to exhaust.
		z, ok := gander.Down(z)
		req.True(ok, "Down should succeed")

		z = gander.Next(z)

		asrt.True(gander.IsEnd(z), "Next after the last node should return the end sentinel")
	})

	t.Run("on end sentinel stays at end sentinel", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		root := ListBranch{Items: []gander.Node{a}}
		z := gander.NewZipper(root)

		// Advance to end sentinel.
		z, ok := gander.Down(z)
		req.True(ok)
		z = gander.Next(z)
		req.True(gander.IsEnd(z), "precondition: should be at end sentinel")

		// Call Next again on the end sentinel.
		z = gander.Next(z)

		asrt.True(gander.IsEnd(z), "Next on end sentinel should remain at end sentinel")
	})

	t.Run("with Replace then Root yields modified tree", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		// Traverse with Next; replace "b" when encountered.
		for !gander.IsEnd(z) {
			if leaf, ok := gander.Focus(z).(StringLeaf); ok && leaf.Value == "b" {
				z, _ = gander.Replace(z, StringLeaf{Value: "B"})
			}
			z = gander.Next(z)
		}

		rootZ, ok := gander.Root(z)
		req.True(ok, "Root on end sentinel should succeed")

		children, ok := gander.Children(rootZ)
		req.True(ok)
		req.Len(children, 3)
		asrt.True(children[0].(StringLeaf).Equal(a), "first child should be a")
		asrt.True(children[1].(StringLeaf).Equal(StringLeaf{Value: "B"}), "second child should be B (replaced)")
		asrt.True(children[2].(StringLeaf).Equal(c), "third child should be c")
	})
}

func TestIsEnd(t *testing.T) {
	t.Run("returns false for normal zipper", func(t *testing.T) {
		asrt := assert.New(t)

		leaf := StringLeaf{Value: "x"}
		z := gander.NewZipper(leaf)

		asrt.False(gander.IsEnd(z), "IsEnd should return false for a freshly created Zipper")
	})

	t.Run("returns true for end sentinel", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		// A single leaf root: traversing past it yields the end sentinel.
		leaf := StringLeaf{Value: "only"}
		z := gander.NewZipper(leaf)

		// Next on root (a leaf) goes directly to end sentinel.
		z = gander.Next(z)
		req.True(gander.IsEnd(z), "after exhausting traversal of a single leaf, IsEnd should be true")

		asrt.True(gander.IsEnd(z), "IsEnd should return true for the end sentinel")
	})
}

func TestPrev(t *testing.T) {
	t.Run("returns false at root", func(t *testing.T) {
		asrt := assert.New(t)

		root := ListBranch{Items: []gander.Node{
			StringLeaf{Value: "a"},
			StringLeaf{Value: "b"},
		}}
		z := gander.NewZipper(root)

		_, ok := gander.Prev(z)
		asrt.False(ok, "Prev at root should return false")
	})

	t.Run("reverses Next", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		// Navigate forward: root → a → b → c
		z = gander.Next(z) // now at a
		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a), "first Next should reach a")

		z = gander.Next(z) // now at b
		focused, ok = gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(b), "second Next should reach b")

		z = gander.Next(z) // now at c
		focused, ok = gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(c), "third Next should reach c")

		// Navigate backward: c → b → a → root
		z, ok = gander.Prev(z) // back to b
		req.True(ok)
		focused, ok = gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(b), "first Prev from c should reach b")

		z, ok = gander.Prev(z) // back to a
		req.True(ok)
		focused, ok = gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a), "second Prev from b should reach a")

		z, ok = gander.Prev(z) // back to root
		req.True(ok)
		rootFocused, ok := gander.Focus(z).(ListBranch)
		req.True(ok)
		asrt.True(rootFocused.Equal(root), "third Prev from a should reach root")
	})

	t.Run("at leftmost child returns parent", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		// Navigate to "a" (leftmost child) via Down.
		z, ok := gander.Down(z)
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a), "precondition: focus should be a")

		// Prev from leftmost child should move to parent.
		z, ok = gander.Prev(z)
		req.True(ok, "Prev from leftmost child should succeed")

		parentFocused, ok := gander.Focus(z).(ListBranch)
		req.True(ok, "focus after Prev from leftmost child should be a ListBranch (the parent)")
		asrt.True(parentFocused.Equal(root), "Prev from leftmost child should focus the parent (root)")
	})

	t.Run("with left sibling returns rightmost descendant of left sibling", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		// Build tree: root → [left → [x, y], right]
		x := StringLeaf{Value: "x"}
		y := StringLeaf{Value: "y"}
		left := ListBranch{Items: []gander.Node{x, y}}
		right := StringLeaf{Value: "right"}
		root := ListBranch{Items: []gander.Node{left, right}}
		z := gander.NewZipper(root)

		// Navigate to "right": Down to left, Right to right.
		z, ok := gander.Down(z) // focus: left
		req.True(ok)
		z, ok = gander.Right(z) // focus: right
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		req.True(focused.Equal(right), "precondition: focus should be right")

		// Prev from "right" should go to "y" (rightmost descendant of left sibling).
		z, ok = gander.Prev(z)
		req.True(ok, "Prev from right should succeed")

		result, ok := gander.Focus(z).(StringLeaf)
		req.True(ok, "focus after Prev should be a StringLeaf")
		asrt.True(result.Equal(y), "Prev from right should focus y (rightmost descendant of left sibling)")
	})
}

// TestRemoveThenNext verifies that Remove followed by Next continues traversal
// correctly, skipping to the next unvisited node rather than re-visiting.
func TestRemoveThenNext(t *testing.T) {
	t.Run("Remove then Next skips to the correct next node", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		// Navigate to "b": Down to a, Right to b.
		z, ok := gander.Down(z) // focus: a
		req.True(ok)
		z, ok = gander.Right(z) // focus: b
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(b), "precondition: focus should be b")

		// Remove "b"; focus moves to the depth-first predecessor (a).
		z, ok = gander.Remove(z)
		req.True(ok)

		afterRemove, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(afterRemove.Equal(a), "after Remove, focus should be on a (depth-first predecessor)")

		// Next from a should move to c, not re-visit b.
		z = gander.Next(z)

		result, ok := gander.Focus(z).(StringLeaf)
		req.True(ok, "focus after Next should be a StringLeaf")
		asrt.True(result.Equal(c), "Next after Remove should move to c, not re-visit b")
	})
}
