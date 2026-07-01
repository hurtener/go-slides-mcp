package autofit

import "github.com/hurtener/go-slides-mcp/internal/contracts"

// rewriteNodes returns a copy of nodes with mutate applied to every node in
// the tree (children before their container, so a container's copy carries
// already-mutated children). Container kinds are rebuilt only when recursing
// into them — a leaf whose mutate call returns the same pointer contributes
// no allocation, so an unaffected slide costs one shallow copy per level.
// A nil input returns nil (absence is preserved, matching the codec).
func rewriteNodes(nodes []contracts.SlideNode, mutate func(contracts.SlideNode) contracts.SlideNode) []contracts.SlideNode {
	if nodes == nil {
		return nil
	}
	out := make([]contracts.SlideNode, len(nodes))
	for i, n := range nodes {
		out[i] = rewriteNode(n, mutate)
	}
	return out
}

// rewriteNode recurses into the closed set of container node kinds the
// ladder cares about (two_column left/right, grid/bento cells, card and
// card_section body — the same shapes Fill's sibling rung inspects) and
// applies mutate bottom-up. Every case builds a fresh shallow copy before
// recursing/mutating, so the input node is never written to. Leaf kinds
// (hero, heading, stat, prose, ...) fall through to the default case, which
// hands the original node straight to mutate — mutate itself is responsible
// for copy-on-write when it changes a leaf.
func rewriteNode(n contracts.SlideNode, mutate func(contracts.SlideNode) contracts.SlideNode) contracts.SlideNode {
	switch v := n.(type) {
	case *contracts.TwoColumn:
		cp := *v
		cp.Left = rewriteNodes(v.Left, mutate)
		cp.Right = rewriteNodes(v.Right, mutate)
		return mutate(&cp)
	case *contracts.Grid:
		cp := *v
		cp.Cells = rewriteNodes(v.Cells, mutate)
		return mutate(&cp)
	case *contracts.Bento:
		return mutate(rewriteBento(v, mutate))
	case *contracts.Card:
		cp := *v
		cp.Body = rewriteNodes(v.Body, mutate)
		return mutate(&cp)
	case *contracts.CardSection:
		cp := *v
		cp.Body = rewriteNodes(v.Body, mutate)
		return mutate(&cp)
	default:
		return mutate(n)
	}
}

// rewriteBento rebuilds a Bento's rows/cells with mutate applied to each
// cell's child node, without touching v.
func rewriteBento(v *contracts.Bento, mutate func(contracts.SlideNode) contracts.SlideNode) *contracts.Bento {
	cp := *v
	if v.Rows == nil {
		return &cp
	}
	cp.Rows = make([]contracts.BentoRow, len(v.Rows))
	for i, row := range v.Rows {
		rowCp := row
		if row.Cells != nil {
			rowCp.Cells = make([]contracts.BentoCell, len(row.Cells))
			for j, cell := range row.Cells {
				cellCp := cell
				cellCp.Node = rewriteNode(cell.Node, mutate)
				rowCp.Cells[j] = cellCp
			}
		}
		cp.Rows[i] = rowCp
	}
	return &cp
}
