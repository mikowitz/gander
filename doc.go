// ABOUTME: Package-level documentation for the gander tree zipper library.
// ABOUTME: This file exists solely to hold the godoc package comment.

// Package gander implements a tree zipper: a functional, immutable cursor for
// navigating and editing arbitrary tree structures.
//
// # Overview
//
// A [Zipper] represents a focused position within a tree. From any position you
// can navigate to parent, child, and sibling nodes, make edits, and produce a
// modified copy of the original tree. Because [Zipper] is a value type, all
// operations return new [Zipper] values — the original is never modified.
//
// The design is based on Gérard Huet's 1997 paper "The Zipper" and closely
// follows the interface of Clojure's clojure.zip library.
//
// # Defining Tree Nodes
//
// To use gander with your own tree types, implement the [Node] interface.
// Embed [BaseNode] to satisfy the unexported marker method without writing it
// by hand. Nodes that can have children must also implement [Branch].
//
//	type MyLeaf struct {
//	    gander.BaseNode
//	    Value string
//	}
//
//	type MyBranch struct {
//	    gander.BaseNode
//	    Items []gander.Node
//	}
//
//	func (b MyBranch) Children() []gander.Node { return b.Items }
//	func (b MyBranch) WithChildren(children []gander.Node) gander.Branch {
//	    return MyBranch{Items: children}
//	}
//
// # Navigation
//
// Create a zipper with [NewZipper], then navigate with [Down], [Up], [Left],
// [Right], [Leftmost], [Rightmost], and [Root]. Most navigation functions
// return (Zipper, bool) where false indicates the move is not possible from
// the current position — for example, [Down] on a leaf or [Left] at the
// leftmost sibling.
//
// # Editing
//
// [Replace] and [Edit] modify the focused node. [InsertLeft], [InsertRight],
// [InsertChild], and [AppendChild] insert new nodes relative to the current
// focus. [Remove] deletes the focused node and moves to its depth-first
// predecessor. Call [Root] to recover the modified root node after any edits.
//
//	z := gander.NewZipper(root)
//	z, _ = gander.Down(z)
//	z, _ = gander.Replace(z, newNode)
//	rootZ, _ := gander.Root(z) // rootZ focuses the modified root
//
// # Depth-First Traversal
//
// [Next] and [Prev] traverse the tree in depth-first order. [IsEnd] detects
// the end sentinel produced when [Next] exhausts the tree. Call [Root] on the
// end sentinel to recover the (possibly modified) root.
//
//	for !gander.IsEnd(z) {
//	    node := gander.Focus(z)
//	    // inspect or conditionally edit node
//	    z = gander.Next(z)
//	}
//	rootZ, _ := gander.Root(z)
package gander
