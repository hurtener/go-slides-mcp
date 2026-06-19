package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/comment"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func (h *handlers) addComment(_ context.Context, in contracts.AddCommentInput) (tool.Result[contracts.AddCommentOutput], error) {
	stored, err := h.deps.Comments.Add(comment.Comment{
		DeckID: in.DeckID,
		Target: comment.Target{Kind: in.Target.Kind, SlideID: in.Target.SlideID, IRPath: append([]any(nil), in.Target.IRPath...)},
		Body:   in.Body,
		Kind:   in.Kind,
		Origin: in.Origin,
	})
	if err != nil {
		return tool.Result[contracts.AddCommentOutput]{}, err
	}
	out := contracts.AddCommentOutput{CommentID: stored.ID}
	return tool.Result[contracts.AddCommentOutput]{Text: fmt.Sprintf("Added comment %q to deck %q.", stored.ID, stored.DeckID), Structured: out}, nil
}

func (h *handlers) listComments(_ context.Context, in contracts.ListCommentsInput) (tool.Result[contracts.ListCommentsOutput], error) {
	stored := h.deps.Comments.List(in.DeckID, in.Resolved, in.TargetKind)
	out := contracts.ListCommentsOutput{Comments: make([]contracts.CommentView, 0, len(stored))}
	for _, item := range stored {
		out.Comments = append(out.Comments, commentView(item))
	}
	return tool.Result[contracts.ListCommentsOutput]{Text: fmt.Sprintf("Found %d comment(s) for deck %q.", len(out.Comments), in.DeckID), Structured: out}, nil
}

func (h *handlers) resolveComment(_ context.Context, in contracts.ResolveCommentInput) (tool.Result[contracts.ResolveCommentOutput], error) {
	stored, err := h.deps.Comments.Resolve(in.CommentID, in.Note)
	if err != nil {
		return tool.Result[contracts.ResolveCommentOutput]{}, mapCommentError(in.CommentID, err)
	}
	out := contracts.ResolveCommentOutput{CommentID: stored.ID, Resolved: stored.Resolved}
	return tool.Result[contracts.ResolveCommentOutput]{Text: fmt.Sprintf("Resolved comment %q.", stored.ID), Structured: out}, nil
}

func commentView(item *comment.Comment) contracts.CommentView {
	return contracts.CommentView{
		CommentID: item.ID,
		DeckID:    item.DeckID,
		Target:    contracts.CommentTarget{Kind: item.Target.Kind, SlideID: item.Target.SlideID, IRPath: append([]any(nil), item.Target.IRPath...)},
		Body:      item.Body,
		Kind:      item.Kind,
		Origin:    item.Origin,
		Resolved:  item.Resolved,
		CreatedAt: item.CreatedAt,
	}
}

func mapCommentError(id string, err error) error {
	if errors.Is(err, comment.ErrNotFound) {
		return fmt.Errorf("comment %q not found: %w", id, err)
	}
	return err
}
