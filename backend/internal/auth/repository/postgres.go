package auth

import (
	"context"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/entity"
	sqlcAuth "github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -source=sqlc/querier.go -destination=mocks/querier_mocks.go -package=mocks

type (
	DB interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		sqlcAuth.DBTX
	}
)

type postgresRepository struct {
	queries sqlcAuth.Querier
	db      DB
}

func NewPostgresRepository(db DB) *postgresRepository {
	return &postgresRepository{
		queries: sqlcAuth.New(db),
		db:      db,
	}
}

func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (entity.User, error) {
	userDb, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return entity.User{}, err
	}

	return entity.User{
		ID:           userDb.ID,
		Email:        userDb.Email,
		PasswordHash: userDb.PasswordHash,
		Role:         userDb.Role,
		CreatedAt:    userDb.CreatedAt.Time,
	}, nil
}

func (r *postgresRepository) CreateUser(ctx context.Context, email string, passwordHash string, role string) (entity.User, error) {
	userDb, err := r.queries.CreateUser(ctx, sqlcAuth.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	})

	if err != nil {
		return entity.User{}, err
	}

	return entity.User{
		ID:           userDb.ID,
		Email:        userDb.Email,
		PasswordHash: userDb.PasswordHash,
		Role:         userDb.Role,
		CreatedAt:    userDb.CreatedAt.Time,
	}, nil
}
