package training

import (
	"context"
	"errors"
	"sync"

	scenario "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
)

var (
	ErrSessionNotFound      = errors.New("training session not found")
	ErrSessionAlreadyExists = errors.New("training session already exists")
)

type SessionRepository interface {
	Create(
		ctx context.Context,
		session *scenario.TrainingSession,
	) error

	GetById(
		ctx context.Context,
		sessionID string,
	) (*scenario.TrainingSession, error)

	Update(
		ctx context.Context,
		session *scenario.TrainingSession,
	) error
}

type IDGenerator interface {
	NewID() string
}

type InMemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]scenario.TrainingSession
}

var _ SessionRepository = (*InMemorySessionRepository)(nil)

func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]scenario.TrainingSession),
	}
}

func (r *InMemorySessionRepository) Create(
	ctx context.Context, session *scenario.TrainingSession,
) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; exists {
		return ErrSessionAlreadyExists
	}

	r.sessions[session.ID] = cloneSession(session)
	return nil
}

func (r *InMemorySessionRepository) GetById(
	ctx context.Context, sessionID string,
) (*scenario.TrainingSession, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	stored, exists := r.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	cloned := cloneSession(&stored)

	return &cloned, nil
}

func (r *InMemorySessionRepository) Update(
	ctx context.Context, session *scenario.TrainingSession,
) error {

	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; !exists {
		return ErrSessionNotFound
	}

	r.sessions[session.ID] = cloneSession(session)
	return nil
}

func cloneSession(
	session *scenario.TrainingSession,
) scenario.TrainingSession {
	cloned := *session

	if session.CompletedAt != nil {
		completedAt := *session.CompletedAt
		cloned.CompletedAt = &completedAt
	}

	return cloned
}
