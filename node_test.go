// ABOUTME: Tests for the core gander interfaces, Zipper construction, and accessor functions.
// ABOUTME: Covers New, Focus, IsBranch, and Children.

package gander_test

import (
	"reflect"
	"testing"

	"github.com/mikowitz/gander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test helper types (used in this and all subsequent test files) ---

// StringLeaf is a simple leaf node holding a string value.
type StringLeaf struct {
	gander.BaseNode
	Value string
}

func (s StringLeaf) Equal(other gander.Node) bool {
	o, ok := other.(StringLeaf)
	return ok && s.Value == o.Value
}

// ListBranch is a simple branch node holding a slice of child nodes.
type ListBranch struct {
	gander.BaseNode
	Items []gander.Node
}

func (lb ListBranch) Children() []gander.Node { return lb.Items }
func (lb ListBranch) WithChildren(children []gander.Node) gander.Branch {
	return ListBranch{Items: children}
}

// Equal uses reflect.DeepEqual because ListBranch contains a slice field
// and is not directly comparable with ==.
func (lb ListBranch) Equal(other gander.Node) bool {
	return reflect.DeepEqual(lb, other)
}

// CountingBranch is a ListBranch variant that tracks WithChildren calls.
// Use it to verify the changed flag optimization: WithChildren should only
// be called when edits have been made in the subtree below this node.
type CountingBranch struct {
	gander.BaseNode
	Items     []gander.Node
	makeCount *int // pointer so the count is shared across value copies
}

func NewCountingBranch(items []gander.Node) CountingBranch {
	n := 0
	return CountingBranch{Items: items, makeCount: &n}
}

func (cb CountingBranch) Children() []gander.Node { return cb.Items }

func (cb CountingBranch) WithChildren(children []gander.Node) gander.Branch {
	*cb.makeCount++
	return CountingBranch{Items: children, makeCount: cb.makeCount}
}

func (cb CountingBranch) MakeCount() int { return *cb.makeCount }

func (cb CountingBranch) Equal(other gander.Node) bool {
	return reflect.DeepEqual(cb, other)
}

// --- Tests ---

var _ gander.Branch = (*ListBranch)(nil)

func TestNew(t *testing.T) {
	t.Run("returns a Zipper with the root node as its focus", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		root := StringLeaf{Value: "root"}
		z := gander.NewZipper(root)

		node := gander.Focus(z)
		req.NotNil(node)

		leaf, ok := node.(StringLeaf)
		req.True(ok, "focused node should be a StringLeaf")
		asrt.Equal("root", leaf.Value)
	})
}

func TestFocus(t *testing.T) {
	t.Run("returns the focused node", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		leaf := StringLeaf{Value: "hello"}
		z := gander.NewZipper(leaf)

		node := gander.Focus(z)
		req.NotNil(node)

		result, ok := node.(StringLeaf)
		req.True(ok, "focused node should be a StringLeaf")
		asrt.True(result.Equal(leaf))
	})
}

func TestIsBranch(t *testing.T) {
	tests := []struct {
		name     string
		node     gander.Node
		expected bool
	}{
		{
			name:     "returns true for a populated ListBranch",
			node:     ListBranch{Items: []gander.Node{StringLeaf{Value: "a"}}},
			expected: true,
		},
		{
			name:     "returns true for an empty ListBranch",
			node:     ListBranch{Items: []gander.Node{}},
			expected: true,
		},
		{
			name:     "returns false for a StringLeaf",
			node:     StringLeaf{Value: "leaf"},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			asrt := assert.New(t)

			z := gander.NewZipper(tc.node)
			asrt.Equal(tc.expected, gander.IsBranch(z))
		})
	}
}

func TestChildren(t *testing.T) {
	t.Run("returns nil and false for a leaf node", func(t *testing.T) {
		asrt := assert.New(t)

		leaf := StringLeaf{Value: "leaf"}
		z := gander.NewZipper(leaf)

		children, ok := gander.Children(z)
		asrt.False(ok, "Children on a leaf should return false")
		asrt.Nil(children, "Children on a leaf should return nil slice")
	})

	t.Run("returns empty slice and true for an empty branch", func(t *testing.T) {
		asrt := assert.New(t)

		branch := ListBranch{Items: []gander.Node{}}
		z := gander.NewZipper(branch)

		children, ok := gander.Children(z)
		asrt.True(ok, "Children on an empty branch should return true")
		asrt.Empty(children, "Children on an empty branch should return an empty slice")
	})

	t.Run("returns the children and true for a populated branch", func(t *testing.T) {
		asrt := assert.New(t)
		req := require.New(t)

		childA := StringLeaf{Value: "a"}
		childB := StringLeaf{Value: "b"}
		branch := ListBranch{Items: []gander.Node{childA, childB}}
		z := gander.NewZipper(branch)

		children, ok := gander.Children(z)
		asrt.True(ok, "Children on a populated branch should return true")
		req.Len(children, 2, "should have exactly two children")

		gotA, ok := children[0].(StringLeaf)
		req.True(ok, "first child should be a StringLeaf")
		asrt.True(gotA.Equal(childA))

		gotB, ok := children[1].(StringLeaf)
		req.True(ok, "second child should be a StringLeaf")
		asrt.True(gotB.Equal(childB))
	})
}
