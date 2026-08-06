package postgres

import (
	"errors"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/domain"
	sqlcChat "github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/repository/postgres/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	_ domain.ChatRepository     = (*ChatRepository)(nil)
	_ domain.MessageRepository  = (*MessageRepository)(nil)
	_ domain.IncidentRepository = (*IncidentRepository)(nil)
)

type DB interface {
	sqlcChat.DBTX
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation
}
