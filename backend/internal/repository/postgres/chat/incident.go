package chatpostgres

import (
	"context"
	"fmt"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
	sqlcChat "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/postgres/chat/sqlc"
)

type IncidentRepository struct {
	queries sqlcChat.Querier
}

func NewIncidentRepository(db DB) *IncidentRepository {
	return &IncidentRepository{queries: sqlcChat.New(db)}
}

func (r *IncidentRepository) ListByChatID(ctx context.Context, chatID int64, cursor chatdomain.Cursor) ([]*chatdomain.Incident, error) {
	if err := cursor.Validate(); err != nil {
		return nil, err
	}

	rows, err := r.queries.ListIncidentsByChatID(ctx, sqlcChat.ListIncidentsByChatIDParams{
		ChatID:  chatID,
		AfterID: cursor.AfterID,
		Lim:     int32(cursor.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents by chat id query: %w", err)
	}

	incidents := make([]*chatdomain.Incident, 0, len(rows))
	for _, row := range rows {
		incidents = append(incidents, &chatdomain.Incident{
			ID:        row.ID,
			ChatID:    row.ChatID,
			Type:      chatdomain.IncidentType(row.Type),
			Comment:   row.Comment,
			CreatedAt: row.CreatedAt,
		})
	}

	return incidents, nil
}

func (r *IncidentRepository) Create(ctx context.Context, incident *chatdomain.Incident, score int64) error {
	row, err := r.queries.CreateIncident(ctx, sqlcChat.CreateIncidentParams{
		ChatID:  incident.ChatID,
		Type:    string(incident.Type),
		Comment: incident.Comment,
		Score:   score,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return chatdomain.ErrChatNotFound
		}
		return fmt.Errorf("failed to create incident query: %w", err)
	}

	incident.ID = row.ID
	incident.CreatedAt = row.CreatedAt

	return nil
}
