package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/comment"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func TestCommentHandlersAddListResolveRoundTrip(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	added, err := h.addComment(ctx, contracts.AddCommentInput{
		DeckID: "deck_123",
		Target: contracts.CommentTarget{Kind: "slide", SlideID: "slide_1"},
		Body:   "Please tighten this slide headline.",
		Kind:   "todo",
		Origin: "agent",
	})
	if err != nil {
		t.Fatalf("addComment: %v", err)
	}
	if added.Structured.CommentID == "" {
		t.Fatal("addComment returned empty comment id")
	}

	listed, err := h.listComments(ctx, contracts.ListCommentsInput{DeckID: "deck_123"})
	if err != nil {
		t.Fatalf("listComments: %v", err)
	}
	if len(listed.Structured.Comments) != 1 {
		t.Fatalf("listComments len = %d, want 1", len(listed.Structured.Comments))
	}
	if listed.Structured.Comments[0].CommentID != added.Structured.CommentID {
		t.Fatalf("listComments comment id = %q, want %q", listed.Structured.Comments[0].CommentID, added.Structured.CommentID)
	}
	if listed.Structured.Comments[0].CreatedAt == "" {
		t.Fatal("listComments returned empty createdAt")
	}

	resolved, err := h.resolveComment(ctx, contracts.ResolveCommentInput{CommentID: added.Structured.CommentID, Note: "Updated in the next revision."})
	if err != nil {
		t.Fatalf("resolveComment: %v", err)
	}
	if !resolved.Structured.Resolved {
		t.Fatal("resolveComment Resolved = false, want true")
	}

	stored := h.deps.Comments.List("deck_123", nil, "")
	if len(stored) != 1 || !stored[0].Resolved || stored[0].ResolveNote != "Updated in the next revision." {
		t.Fatalf("store after resolve = %+v", stored)
	}
	resolvedOnly := true
	listedResolved, err := h.listComments(ctx, contracts.ListCommentsInput{DeckID: "deck_123", Resolved: &resolvedOnly})
	if err != nil {
		t.Fatalf("listComments resolved: %v", err)
	}
	if len(listedResolved.Structured.Comments) != 1 {
		t.Fatalf("listComments resolved len = %d, want 1", len(listedResolved.Structured.Comments))
	}
}

func TestCommentHandlersFilterByResolvedAndTargetKind(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	deckComment, err := h.addComment(ctx, contracts.AddCommentInput{DeckID: "deck_123", Target: contracts.CommentTarget{Kind: "deck"}, Body: "Deck-level note"})
	if err != nil {
		t.Fatalf("add deck comment: %v", err)
	}
	if _, err := h.addComment(ctx, contracts.AddCommentInput{DeckID: "deck_123", Target: contracts.CommentTarget{Kind: "node", SlideID: "slide_1", IRPath: []any{"nodes", 0}}, Body: "Node-level question", Kind: "question", Origin: "app"}); err != nil {
		t.Fatalf("add node comment: %v", err)
	}
	if _, err := h.resolveComment(ctx, contracts.ResolveCommentInput{CommentID: deckComment.Structured.CommentID, Note: "Handled."}); err != nil {
		t.Fatalf("resolve deck comment: %v", err)
	}

	resolvedOnly := true
	resolvedComments, err := h.listComments(ctx, contracts.ListCommentsInput{DeckID: "deck_123", Resolved: &resolvedOnly})
	if err != nil {
		t.Fatalf("list resolved comments: %v", err)
	}
	if len(resolvedComments.Structured.Comments) != 1 || resolvedComments.Structured.Comments[0].Target.Kind != "deck" {
		t.Fatalf("resolved comments = %+v", resolvedComments.Structured.Comments)
	}

	nodeComments, err := h.listComments(ctx, contracts.ListCommentsInput{DeckID: "deck_123", TargetKind: "node"})
	if err != nil {
		t.Fatalf("list node comments: %v", err)
	}
	if len(nodeComments.Structured.Comments) != 1 || nodeComments.Structured.Comments[0].Target.Kind != "node" {
		t.Fatalf("node comments = %+v", nodeComments.Structured.Comments)
	}
	if got := nodeComments.Structured.Comments[0].Target.IRPath; len(got) != 2 {
		t.Fatalf("node comment irPath len = %d, want 2", len(got))
	}
}

func TestResolveCommentMissingReturnsNotFound(t *testing.T) {
	h := testHandlers()
	if _, err := h.resolveComment(context.Background(), contracts.ResolveCommentInput{CommentID: "cmt_missing"}); !errors.Is(err, comment.ErrNotFound) {
		t.Fatalf("resolveComment missing err = %v, want ErrNotFound", err)
	}
}
