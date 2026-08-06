package chatpostgres

import (
	"context"
	"fmt"

	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
	sqlcChat "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/postgres/chat/sqlc"
)

type DecisionRepository struct {
	queries sqlcChat.Querier
}

func NewDecisionRepository(db DB) *DecisionRepository {
	return &DecisionRepository{queries: sqlcChat.New(db)}
}

func (r *DecisionRepository) ListByChatID(
	ctx context.Context,
	chatID int64,
	cursor chatdomain.Cursor,
) ([]*chatdomain.Decision, error) {
	if err := cursor.Validate(); err != nil {
		return nil, err
	}

	rows, err := r.queries.ListDecisionsByChatID(ctx, sqlcChat.ListDecisionsByChatIDParams{
		ChatID:  chatID,
		AfterID: cursor.AfterID,
		Lim:     int32(cursor.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list decisions by chat id query: %w", err)
	}

	decisions := make([]*chatdomain.Decision, 0, len(rows))
	for _, row := range rows {
		decisions = append(decisions, &chatdomain.Decision{
			ID:           row.ID,
			ChatID:       row.ChatID,
			NodeID:       row.NodeID,
			TransitionID: row.TransitionID,
			TargetNodeID: row.TargetNodeID,
			ScoreDelta:   int(row.ScoreDelta),
			Feedback:     row.Feedback,
			CreatedAt:    row.CreatedAt,
		})
	}

	return decisions, nil
}

// Create writes the decision and the resulting chat score in a single
// statement, so the journal and the score can never drift apart.
func (r *DecisionRepository) Create(
	ctx context.Context,
	decision *chatdomain.Decision,
	score int64,
) error {
	row, err := r.queries.CreateDecision(ctx, sqlcChat.CreateDecisionParams{
		ChatID:       decision.ChatID,
		NodeID:       decision.NodeID,
		TransitionID: decision.TransitionID,
		TargetNodeID: decision.TargetNodeID,
		ScoreDelta:   int64(decision.ScoreDelta),
		Feedback:     decision.Feedback,
		Score:        score,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return chatdomain.ErrChatNotFound
		}
		return fmt.Errorf("failed to create decision query: %w", err)
	}

	decision.ID = row.ID
	decision.CreatedAt = row.CreatedAt

	return nil
}
