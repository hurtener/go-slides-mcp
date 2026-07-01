package contracts

// Tree is a hierarchy / org-chart / taxonomy node (R14.10, D-127): a root
// with children laid out as a balanced top-down (or left-right) tidy tree,
// with elbow connector edges between parent and children and soul-styled
// nodes. Native shapes — no asset. Mirror of pptx-go's scene.Tree. Pure
// integer-EMU layout → byte-identical across renders/worker counts;
// depth/breadth past the safe area clamp + warn at render time. A deck with
// no Tree is byte-identical (a new node, absent until used).
type Tree struct {
	// Root is the tree's root node; its Children recurse to build the full
	// hierarchy.
	Root TreeNode `json:"root"`
	// Orientation selects a top-down (vertical, default) or left-right
	// (horizontal) layout. Reuses the shared FlowOrientation enum.
	Orientation FlowOrientation `json:"orientation,omitempty"`
}

func (Tree) slideNodeKind() Kind { return KindTree }

// MarshalJSON injects the "tree" kind discriminator via marshalNode.
func (t *Tree) MarshalJSON() ([]byte, error) { return marshalNode(KindTree, *t) }

func init() { registerNodeKind(KindTree, func() SlideNode { return &Tree{} }) }

// TreeNode is one node in a Tree (D-127): a label + optional detail/icon,
// child nodes, and an AccentIndex selecting its border color from a pinned
// token cycle. Children are concrete TreeNode values (not the SlideNode
// interface), so the default JSON unmarshal recurses without a custom
// UnmarshalJSON.
type TreeNode struct {
	// Label is the node's headline text.
	Label string `json:"label,omitempty"`
	// Detail is optional supporting text under the label.
	Detail string `json:"detail,omitempty"`
	// Icon is an optional curated icon name drawn inside the node.
	Icon string `json:"icon,omitempty"`
	// Children are this node's child nodes, recursing to build the tree.
	Children []TreeNode `json:"children,omitempty"`
	// AccentIndex selects a soul-driven series accent color for the node's
	// border (0 = the first accent). A plain int passthrough — 0 is a real
	// value (the first accent), not "unset".
	AccentIndex int `json:"accentIndex,omitempty"`
}
