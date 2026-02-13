// Package agents provides AI agent implementations for FutureBuild.
// This file defines the FeedWriter interface for agents to write feed cards.
// See FRONTEND_V2_SPEC.md §6.2
package agents

import (
	"context"

	"github.com/colton/futurebuild/internal/models"
)

// FeedWriter allows agents to write cards into the portfolio feed.
// Thin interface — agents only write, never read. Reading is done by the API layer.
// Satisfied by *service.FeedService.
type FeedWriter interface {
	WriteCard(ctx context.Context, card *models.FeedCard) error
}
