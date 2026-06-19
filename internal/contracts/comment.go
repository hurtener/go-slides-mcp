package contracts

// CommentTarget is the typed comment target address.
type CommentTarget struct {
	// Kind identifies whether the comment points at a deck, slide, or node.
	Kind string `json:"kind"`
	// SlideID identifies the target slide when the comment is slide- or node-scoped.
	SlideID string `json:"slideId,omitempty"`
	// IRPath is the optional structural path to a node target.
	IRPath []any `json:"irPath,omitempty"`
}

// AddCommentInput is the typed input for add_comment.
type AddCommentInput struct {
	// DeckID identifies the deck that owns the comment thread.
	DeckID string `json:"deckId"`
	// Target identifies what the comment points at.
	Target CommentTarget `json:"target"`
	// Body is the comment text.
	Body string `json:"body"`
	// Kind is the optional comment classification such as note, todo, or question.
	Kind string `json:"kind,omitempty"`
	// Origin is the optional caller origin such as agent or app.
	Origin string `json:"origin,omitempty"`
}

// AddCommentOutput is the structured result for add_comment.
type AddCommentOutput struct {
	// CommentID is the stored comment ID.
	CommentID string `json:"commentId"`
}

// ListCommentsInput is the typed input for list_comments.
type ListCommentsInput struct {
	// DeckID identifies which deck's comments to list.
	DeckID string `json:"deckId"`
	// Resolved optionally filters comments by resolution state.
	Resolved *bool `json:"resolved,omitempty"`
	// TargetKind optionally filters comments by target kind.
	TargetKind string `json:"targetKind,omitempty"`
}

// CommentView is the structured metadata for one stored comment.
type CommentView struct {
	// CommentID is the stored comment ID.
	CommentID string `json:"commentId"`
	// DeckID identifies the deck that owns the comment.
	DeckID string `json:"deckId"`
	// Target identifies what the comment points at.
	Target CommentTarget `json:"target"`
	// Body is the comment text.
	Body string `json:"body"`
	// Kind is the comment classification.
	Kind string `json:"kind,omitempty"`
	// Origin is the caller origin such as agent or app.
	Origin string `json:"origin,omitempty"`
	// Resolved reports whether the comment has been resolved.
	Resolved bool `json:"resolved"`
	// CreatedAt is the comment creation time in UTC RFC3339 format.
	CreatedAt string `json:"createdAt"`
}

// ListCommentsOutput is the structured result for list_comments.
type ListCommentsOutput struct {
	// Comments is the ordered list of comments that matched the filters.
	Comments []CommentView `json:"comments,omitempty"`
}

// ResolveCommentInput is the typed input for resolve_comment.
type ResolveCommentInput struct {
	// CommentID identifies which stored comment to resolve.
	CommentID string `json:"commentId"`
	// Note is the optional resolution note.
	Note string `json:"note,omitempty"`
}

// ResolveCommentOutput is the structured result for resolve_comment.
type ResolveCommentOutput struct {
	// CommentID identifies the resolved comment.
	CommentID string `json:"commentId"`
	// Resolved reports whether the comment is now resolved.
	Resolved bool `json:"resolved"`
}
