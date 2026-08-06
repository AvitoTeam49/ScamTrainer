package chatpostgres

import (
	"errors"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
	sqlcChat "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/postgres/chat/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	_ chatdomain.ChatRepository     = (*ChatRepository)(nil)
	_ chatdomain.MessageRepository  = (*MessageRepository)(nil)
	_ chatdomain.DecisionRepository = (*DecisionRepository)(nil)
)

type DB interface {
	sqlcChat.DBTX
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation
}
