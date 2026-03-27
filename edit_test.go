// ABOUTME: Tests for Replace, Edit, and Root editing operations (Step 5).
// ABOUTME: Verifies focus replacement, tree reconstruction, and the changed-flag optimization.

package gander_test

import (
	"testing"

	"github.com/mikowitz/gander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplace(t *testing.T) {
	t.Run("changes the focused node and Focus reflects it", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		original := StringLeaf{Value: "original"}
		replacement := StringLeaf{Value: "replacement"}
		z := gander.NewZipper(original)

		z, ok := gander.Replace(z, replacement)
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(replacement))
	})

	t.Run("then Up then Focus shows reconstructed parent with new child", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		c := StringLeaf{Value: "c"}
		root := ListBranch{Items: []gander.Node{a, b, c}}
		z := gander.NewZipper(root)

		// Navigate down to the first child, replace it, then go up.
		z, ok := gander.Down(z)
		req.True(ok)

		replacement := StringLeaf{Value: "z"}
		z, ok = gander.Replace(z, replacement)
		req.True(ok)

		z, ok = gander.Up(z)
		req.True(ok)

		// The parent should now contain the replacement as its first child.
		children, ok := gander.Children(z)
		req.True(ok)
		req.Len(children, 3)

		first, ok := children[0].(StringLeaf)
		req.True(ok)
		asrt.True(first.Equal(replacement), "first child should be the replacement")

		second, ok := children[1].(StringLeaf)
		req.True(ok)
		asrt.True(second.Equal(b), "second child should be unchanged")

		third, ok := children[2].(StringLeaf)
		req.True(ok)
		asrt.True(third.Equal(c), "third child should be unchanged")
	})
}

func TestRoot(t *testing.T) {
	t.Run("Replace then Root returns the modified tree", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		// Replace the first child and zip all the way back to root.
		z, ok := gander.Down(z)
		req.True(ok)

		replacement := StringLeaf{Value: "x"}
		z, ok = gander.Replace(z, replacement)
		req.True(ok)

		z, ok = gander.Root(z)
		req.True(ok)

		// Root focus should be a ListBranch with the replaced child.
		rootFocused, ok := gander.Focus(z).(ListBranch)
		req.True(ok)

		expected := ListBranch{Items: []gander.Node{replacement, b}}
		asrt.True(rootFocused.Equal(expected))
	})

	t.Run("Replace deep in tree then Root propagates changes all the way up", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		// Build a tree: root → [inner → [leaf_a, leaf_b], leaf_c]
		leafA := StringLeaf{Value: "a"}
		leafB := StringLeaf{Value: "b"}
		leafC := StringLeaf{Value: "c"}
		inner := ListBranch{Items: []gander.Node{leafA, leafB}}
		root := ListBranch{Items: []gander.Node{inner, leafC}}
		z := gander.NewZipper(root)

		// Navigate to leaf_a (two levels deep) and replace it.
		z, ok := gander.Down(z) // focus: inner
		req.True(ok)

		z, ok = gander.Down(z) // focus: leaf_a
		req.True(ok)

		replacement := StringLeaf{Value: "z"}
		z, ok = gander.Replace(z, replacement)
		req.True(ok)

		z, ok = gander.Root(z)
		req.True(ok)

		// The root should reflect the deep replacement.
		rootFocused, ok := gander.Focus(z).(ListBranch)
		req.True(ok)

		expectedInner := ListBranch{Items: []gander.Node{replacement, leafB}}
		expectedRoot := ListBranch{Items: []gander.Node{expectedInner, leafC}}
		asrt.True(rootFocused.Equal(expectedRoot))
	})
}

func TestEdit(t *testing.T) {
	t.Run("applies function to current node and replaces focus", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		original := StringLeaf{Value: "hello"}
		z := gander.NewZipper(original)

		transform := func(n gander.Node) gander.Node {
			leaf := n.(StringLeaf)
			return StringLeaf{Value: leaf.Value + "_edited"}
		}

		z, ok := gander.Edit(z, transform)
		req.True(ok)

		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(StringLeaf{Value: "hello_edited"}))
	})

	t.Run("Edit then Root reflects the transformation in the full tree", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		z, ok := gander.Down(z)
		req.True(ok)

		// Transform leaf "a" to "A" (the original value is not used in the output).
		z, ok = gander.Edit(z, func(n gander.Node) gander.Node {
			_ = n.(StringLeaf)
			return StringLeaf{Value: "A"}
		})
		req.True(ok)

		z, ok = gander.Root(z)
		req.True(ok)

		rootFocused, ok := gander.Focus(z).(ListBranch)
		req.True(ok)

		expected := ListBranch{Items: []gander.Node{StringLeaf{Value: "A"}, b}}
		asrt.True(rootFocused.Equal(expected))
	})
}

// TestInsertLeftAndInsertRightAtRoot verifies that both InsertLeft and InsertRight return
// false when called at the root, since root has no sibling slots.
func TestInsertLeftAndInsertRightAtRoot(t *testing.T) {
	tests := []struct {
		name   string
		insert func(gander.Zipper, gander.Node) (gander.Zipper, bool)
	}{
		{
			name:   "InsertLeft at root returns false",
			insert: gander.InsertLeft,
		},
		{
			name:   "InsertRight at root returns false",
			insert: gander.InsertRight,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			asrt := assert.New(t)
			req := require.New(t)

			root := ListBranch{Items: []gander.Node{
				StringLeaf{Value: "a"},
				StringLeaf{Value: "b"},
			}}
			z := gander.NewZipper(root)

			_, ok := tc.insert(z, StringLeaf{Value: "new"})
			req.False(ok)
			// Focus must remain unchanged after a failed insert.
			focused, ok := gander.Focus(z).(ListBranch)
			req.True(ok)
			asrt.True(focused.Equal(root))
		})
	}
}

// TestInsertLeft verifies InsertLeft behavior: focus is unchanged, left sibling list grows,
// the inserted node appears in the reconstructed tree, and navigation reaches it.
func TestInsertLeft(t *testing.T) {
	t.Run("adds to left siblings and focus stays on current node", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		// Navigate down to "b" so it has a left sibling.
		z, ok := gander.Down(z) // focus: "a"
		req.True(ok)
		z, ok = gander.Right(z) // focus: "b"
		req.True(ok)

		newNode := StringLeaf{Value: "inserted"}
		z, ok = gander.InsertLeft(z, newNode)
		req.True(ok)

		// Focus must still be "b".
		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(b), "focus must remain on b after InsertLeft")

		// Lefts must now contain both "a" and the inserted node (in tree order: a, inserted).
		lefts := gander.Lefts(z)
		req.Len(lefts, 2, "should have 2 left siblings after InsertLeft")
		asrt.True(lefts[0].(StringLeaf).Equal(a), "first left sibling should be a")
		asrt.True(lefts[1].(StringLeaf).Equal(newNode), "second left sibling should be the inserted node")
	})

	t.Run("Root shows the inserted node in the tree and Left navigates to it", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		// Navigate down to "b", insert a sibling to its left.
		z, ok := gander.Down(z) // focus: "a"
		req.True(ok)
		z, ok = gander.Right(z) // focus: "b"
		req.True(ok)

		newNode := StringLeaf{Value: "between"}
		z, ok = gander.InsertLeft(z, newNode)
		req.True(ok)

		// Zip up to root and verify the full children list: [a, between, b].
		rootZ, ok := gander.Root(z)
		req.True(ok)

		children, ok := gander.Children(rootZ)
		req.True(ok)
		req.Len(children, 3, "root should now have 3 children")
		asrt.True(children[0].(StringLeaf).Equal(a), "first child should be a")
		asrt.True(children[1].(StringLeaf).Equal(newNode), "second child should be the inserted node")
		asrt.True(children[2].(StringLeaf).Equal(b), "third child should be b")

		// Navigate left from "b" to reach the inserted node.
		prev, ok := gander.Left(z)
		req.True(ok)
		prevFocused, ok := gander.Focus(prev).(StringLeaf)
		req.True(ok)
		asrt.True(prevFocused.Equal(newNode), "Left from b should land on the inserted node")
	})
}

// TestInsertRight verifies InsertRight behavior: focus is unchanged, right sibling list grows,
// the inserted node appears in the reconstructed tree, and navigation reaches it.
func TestInsertRight(t *testing.T) {
	t.Run("adds to right siblings and focus stays on current node", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		// Navigate down to "a"; it has "b" as a right sibling.
		z, ok := gander.Down(z) // focus: "a"
		req.True(ok)

		newNode := StringLeaf{Value: "inserted"}
		z, ok = gander.InsertRight(z, newNode)
		req.True(ok)

		// Focus must still be "a".
		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a), "focus must remain on a after InsertRight")

		// Rights must now contain the inserted node first, then "b".
		rights := gander.Rights(z)
		req.Len(rights, 2, "should have 2 right siblings after InsertRight")
		asrt.True(rights[0].(StringLeaf).Equal(newNode), "first right sibling should be the inserted node")
		asrt.True(rights[1].(StringLeaf).Equal(b), "second right sibling should be b")
	})

	t.Run("Root shows the inserted node in the tree and Right navigates to it", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		// Navigate down to "a", insert a sibling to its right.
		z, ok := gander.Down(z) // focus: "a"
		req.True(ok)

		newNode := StringLeaf{Value: "between"}
		z, ok = gander.InsertRight(z, newNode)
		req.True(ok)

		// Zip up to root and verify the full children list: [a, between, b].
		rootZ, ok := gander.Root(z)
		req.True(ok)

		children, ok := gander.Children(rootZ)
		req.True(ok)
		req.Len(children, 3, "root should now have 3 children")
		asrt.True(children[0].(StringLeaf).Equal(a), "first child should be a")
		asrt.True(children[1].(StringLeaf).Equal(newNode), "second child should be the inserted node")
		asrt.True(children[2].(StringLeaf).Equal(b), "third child should be b")

		// Navigate right from "a" to reach the inserted node.
		next, ok := gander.Right(z)
		req.True(ok)
		nextFocused, ok := gander.Focus(next).(StringLeaf)
		req.True(ok)
		asrt.True(nextFocused.Equal(newNode), "Right from a should land on the inserted node")
	})
}

// TestInsertAccumulation verifies that multiple consecutive InsertLeft and InsertRight calls
// each accumulate correctly in the sibling lists.
func TestInsertAccumulation(t *testing.T) {
	t.Run("multiple InsertLeft calls accumulate in correct tree order", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		// Navigate to "b" and insert two nodes to its left.
		z, ok := gander.Down(z) // focus: "a"
		req.True(ok)
		z, ok = gander.Right(z) // focus: "b"
		req.True(ok)

		first := StringLeaf{Value: "first"}
		second := StringLeaf{Value: "second"}

		z, ok = gander.InsertLeft(z, first)
		req.True(ok)
		z, ok = gander.InsertLeft(z, second)
		req.True(ok)

		// Focus must still be "b".
		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(b))

		// Root should show [a, first, second, b] in tree order.
		rootZ, ok := gander.Root(z)
		req.True(ok)

		children, ok := gander.Children(rootZ)
		req.True(ok)
		req.Len(children, 4)
		asrt.True(children[0].(StringLeaf).Equal(a), "children[0] should be a")
		asrt.True(children[1].(StringLeaf).Equal(first), "children[1] should be first")
		asrt.True(children[2].(StringLeaf).Equal(second), "children[2] should be second")
		asrt.True(children[3].(StringLeaf).Equal(b), "children[3] should be b")
	})

	t.Run("multiple InsertRight calls accumulate in correct tree order", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		// Navigate to "a" and insert two nodes to its right.
		z, ok := gander.Down(z) // focus: "a"
		req.True(ok)

		first := StringLeaf{Value: "first"}
		second := StringLeaf{Value: "second"}

		z, ok = gander.InsertRight(z, first)
		req.True(ok)
		z, ok = gander.InsertRight(z, second)
		req.True(ok)

		// Focus must still be "a".
		focused, ok := gander.Focus(z).(StringLeaf)
		req.True(ok)
		asrt.True(focused.Equal(a))

		// Root should show [a, second, first, b] in tree order because each InsertRight
		// prepends to the right sibling list, so the last-inserted node is nearest.
		rootZ, ok := gander.Root(z)
		req.True(ok)

		children, ok := gander.Children(rootZ)
		req.True(ok)
		req.Len(children, 4)
		asrt.True(children[0].(StringLeaf).Equal(a), "children[0] should be a")
		asrt.True(children[1].(StringLeaf).Equal(second), "children[1] should be second (nearest right)")
		asrt.True(children[2].(StringLeaf).Equal(first), "children[2] should be first")
		asrt.True(children[3].(StringLeaf).Equal(b), "children[3] should be b")
	})
}

// TestInsertChildAndAppendChildOnLeaf verifies that both InsertChild and AppendChild return
// false when called on a leaf node, since leaves cannot have children.
func TestInsertChildAndAppendChildOnLeaf(t *testing.T) {
	tests := []struct {
		name   string
		insert func(gander.Zipper, gander.Node) (gander.Zipper, bool)
	}{
		{
			name:   "InsertChild on a leaf returns false",
			insert: gander.InsertChild,
		},
		{
			name:   "AppendChild on a leaf returns false",
			insert: gander.AppendChild,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			asrt := assert.New(t)
			req := require.New(t)

			leaf := StringLeaf{Value: "leaf"}
			z := gander.NewZipper(leaf)

			_, ok := tc.insert(z, StringLeaf{Value: "child"})
			req.False(ok)
			// Focus must remain unchanged after a failed insert.
			focused, ok := gander.Focus(z).(StringLeaf)
			req.True(ok)
			asrt.True(focused.Equal(leaf))
		})
	}
}

// TestInsertChild verifies InsertChild behavior: focus stays on parent, inserted node becomes
// leftmost child, Down navigates to it, and changes propagate to Root.
func TestInsertChild(t *testing.T) {
	t.Run("adds as leftmost child, focus stays on parent, Down focuses inserted child, Root propagates changes", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		newChild := StringLeaf{Value: "leftmost"}
		z, ok := gander.InsertChild(z, newChild)
		req.True(ok)

		// Focus must still be the root branch.
		focused, ok := gander.Focus(z).(ListBranch)
		req.True(ok)
		// The parent focus is the original root value; the new child is visible via Down/Root.
		_ = focused

		// Down from the parent must land on the inserted (leftmost) child.
		child, ok := gander.Down(z)
		req.True(ok)
		childFocused, ok := gander.Focus(child).(StringLeaf)
		req.True(ok)
		asrt.True(childFocused.Equal(newChild), "Down from parent should focus the inserted leftmost child")

		// Root must show the inserted node as the first child.
		rootZ, ok := gander.Root(z)
		req.True(ok)
		children, ok := gander.Children(rootZ)
		req.True(ok)
		req.Len(children, 3, "root should now have 3 children")
		asrt.True(children[0].(StringLeaf).Equal(newChild), "first child should be the inserted node")
		asrt.True(children[1].(StringLeaf).Equal(a), "second child should be a")
		asrt.True(children[2].(StringLeaf).Equal(b), "third child should be b")
	})
}

// TestAppendChild verifies AppendChild behavior: focus stays on parent, inserted node becomes
// rightmost child (including on an empty branch), and changes propagate to Root.
func TestAppendChild(t *testing.T) {
	t.Run("adds as rightmost child, focus stays on parent, and Root propagates changes", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		a := StringLeaf{Value: "a"}
		b := StringLeaf{Value: "b"}
		root := ListBranch{Items: []gander.Node{a, b}}
		z := gander.NewZipper(root)

		newChild := StringLeaf{Value: "rightmost"}
		z, ok := gander.AppendChild(z, newChild)
		req.True(ok)

		// Focus must still be the root branch.
		focused, ok := gander.Focus(z).(ListBranch)
		req.True(ok)
		_ = focused

		// Root must show the appended node as the last child.
		rootZ, ok := gander.Root(z)
		req.True(ok)
		children, ok := gander.Children(rootZ)
		req.True(ok)
		req.Len(children, 3, "root should now have 3 children")
		asrt.True(children[0].(StringLeaf).Equal(a), "first child should be a")
		asrt.True(children[1].(StringLeaf).Equal(b), "second child should be b")
		asrt.True(children[2].(StringLeaf).Equal(newChild), "third child should be the appended node")
	})

	t.Run("on empty branch creates a single child", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		empty := ListBranch{Items: []gander.Node{}}
		z := gander.NewZipper(empty)

		onlyChild := StringLeaf{Value: "only"}
		z, ok := gander.AppendChild(z, onlyChild)
		req.True(ok)

		// Root shows the single appended child.
		rootZ, ok := gander.Root(z)
		req.True(ok)
		children, ok := gander.Children(rootZ)
		req.True(ok)
		req.Len(children, 1, "branch should now have exactly one child")
		asrt.True(children[0].(StringLeaf).Equal(onlyChild), "the single child should be the appended node")
	})
}

// TestUpReconstructionOptimization verifies the changed-flag behavior via CountingBranch.
// Two cases are tested in a table: one where no edit is made (MakeCount stays 0) and one
// where Replace is called before Up (MakeCount becomes 1).
func TestUpReconstructionOptimization(t *testing.T) {
	tests := []struct {
		name          string
		editChild     bool // if true, call Replace on child before Up
		wantMakeCount int
	}{
		{
			name:          "unmodified Up does not call WithChildren",
			editChild:     false,
			wantMakeCount: 0,
		},
		{
			name:          "modified Up calls WithChildren exactly once",
			editChild:     true,
			wantMakeCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			asrt := assert.New(t)
			req := require.New(t)

			cb := NewCountingBranch([]gander.Node{
				StringLeaf{Value: "a"},
				StringLeaf{Value: "b"},
			})
			z := gander.NewZipper(cb)

			z, ok := gander.Down(z)
			req.True(ok)

			if tc.editChild {
				replacement := StringLeaf{Value: "z"}
				z, ok = gander.Replace(z, replacement)
				req.True(ok)
			}

			_, ok = gander.Up(z)
			req.True(ok)

			asrt.Equal(tc.wantMakeCount, cb.MakeCount(),
				"WithChildren call count mismatch for case %q", tc.name)
		})
	}
}
