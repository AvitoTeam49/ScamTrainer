package training

import (
	"context"
	"errors"
	"sync"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/scenario"
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

	GetByID(
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

func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]scenario.TrainingSession),
	}
}

func (r *InMemorySessionRepository) Create(
	ctx context.Context, session *scenario.TrainingSession,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; exists {
		return ErrSessionAlreadyExists
	}

	r.sessions[session.ID] = *session
	return nil
}

func (r *InMemorySessionRepository) GetByID(
	ctx context.Context, sessionID string,
) (*scenario.TrainingSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return &session, nil
}

func (r *InMemorySessionRepository) Update(
	ctx context.Context, session *scenario.TrainingSession,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; !exists {
		return ErrSessionNotFound
	}

	r.sessions[session.ID] = *session
	return nil
}
