package training

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/scenario"
)

func TestInMemorySessionRepository_CreateAndGetById(
	t *testing.T,
) {
	ctx := context.Background()
	repository := NewInMemorySessionRepository()

	session := newTestSession("session-1")

	if err := repository.Create(ctx, session); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	loaded, err := repository.GetById(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}

	requireSessionsEqual(t, loaded, session)
}

func TestInMemorySessionRepository_CreateRejectsDuplicateID(
	t *testing.T,
) {
	ctx := context.Background()
	repository := NewInMemorySessionRepository()

	session := newTestSession("session-1")

	if err := repository.Create(ctx, session); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	err := repository.Create(ctx, session)

	if !errors.Is(err, ErrSessionAlreadyExists) {
		t.Fatalf(
			"second Create() error = %v, want %v",
			err,
			ErrSessionAlreadyExists,
		)
	}
}

func TestInMemorySessionRepository_GetByIdReturnsNotFound(
	t *testing.T,
) {
	ctx := context.Background()
	repository := NewInMemorySessionRepository()

	session, err := repository.GetById(
		ctx,
		"missing-session",
	)

	if session != nil {
		t.Fatalf(
			"GetById() session = %#v, want nil",
			session,
		)
	}

	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf(
			"GetById() error = %v, want %v",
			err,
			ErrSessionNotFound,
		)
	}
}

func TestInMemorySessionRepository_Update(
	t *testing.T,
) {
	ctx := context.Background()
	repository := NewInMemorySessionRepository()

	initial := newTestSession("session-1")

	if err := repository.Create(ctx, initial); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	completedAt := initial.UpdatedAt.Add(time.Minute)

	updated := *initial
	updated.CurrentNodeID = "safe-ending"
	updated.Status = scenario.SessionStatusCompleted
	updated.Score = 20
	updated.UpdatedAt = completedAt
	updated.CompletedAt = &completedAt

	if err := repository.Update(ctx, &updated); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	expected := updated
	expectedCompletedAt := *updated.CompletedAt
	expected.CompletedAt = &expectedCompletedAt

	updated.Score = 999
	tamperedTime := expectedCompletedAt.Add(time.Hour)
	*updated.CompletedAt = tamperedTime

	loaded, err := repository.GetById(ctx, initial.ID)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}

	requireSessionsEqual(t, loaded, &expected)
}

func TestInMemorySessionRepository_UpdateReturnsNotFound(
	t *testing.T,
) {
	ctx := context.Background()
	repository := NewInMemorySessionRepository()

	session := newTestSession("missing-session")

	err := repository.Update(ctx, session)

	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf(
			"Update() error = %v, want %v",
			err,
			ErrSessionNotFound,
		)
	}
}

func TestInMemorySessionRepository_CreateStoresCopy(
	t *testing.T,
) {
	ctx := context.Background()
	repository := NewInMemorySessionRepository()

	completedAt := time.Date(
		2026,
		time.August,
		6,
		12,
		30,
		0,
		0,
		time.UTC,
	)

	session := newTestSession("session-1")
	session.Status = scenario.SessionStatusCompleted
	session.Score = 20
	session.CompletedAt = &completedAt

	expectedCompletedAt := completedAt

	if err := repository.Create(ctx, session); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	session.Score = 999
	tamperedTime := expectedCompletedAt.Add(time.Hour)
	*session.CompletedAt = tamperedTime

	loaded, err := repository.GetById(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}

	if loaded.Score != 20 {
		t.Errorf(
			"stored Score = %d, want 20",
			loaded.Score,
		)
	}

	if loaded.CompletedAt == nil {
		t.Fatal("stored CompletedAt = nil")
	}

	if !loaded.CompletedAt.Equal(expectedCompletedAt) {
		t.Errorf(
			"stored CompletedAt = %v, want %v",
			loaded.CompletedAt,
			expectedCompletedAt,
		)
	}
}

func TestInMemorySessionRepository_GetByIdReturnsCopy(
	t *testing.T,
) {
	ctx := context.Background()
	repository := NewInMemorySessionRepository()

	completedAt := time.Date(
		2026,
		time.August,
		6,
		12,
		30,
		0,
		0,
		time.UTC,
	)

	session := newTestSession("session-1")
	session.Status = scenario.SessionStatusCompleted
	session.Score = 20
	session.CompletedAt = &completedAt

	expectedCompletedAt := completedAt

	if err := repository.Create(ctx, session); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	loaded, err := repository.GetById(ctx, session.ID)
	if err != nil {
		t.Fatalf("first GetByID() error = %v", err)
	}

	loaded.Score = 999
	tamperedTime := expectedCompletedAt.Add(time.Hour)
	*loaded.CompletedAt = tamperedTime

	stored, err := repository.GetById(ctx, session.ID)
	if err != nil {
		t.Fatalf("second GetById() error = %v", err)
	}

	if stored.Score != 20 {
		t.Errorf(
			"stored Score = %d, want 20",
			stored.Score,
		)
	}

	if stored.CompletedAt == nil {
		t.Fatal("stored CompletedAt = nil")
	}

	if !stored.CompletedAt.Equal(expectedCompletedAt) {
		t.Errorf(
			"stored CompletedAt = %v, want %v",
			stored.CompletedAt,
			expectedCompletedAt,
		)
	}
}

func TestInMemorySessionRepository_ConcurrentAccess(
	t *testing.T,
) {
	const sessionCount = 100

	ctx := context.Background()
	repository := NewInMemorySessionRepository()

	var waitGroup sync.WaitGroup

	start := make(chan struct{})

	errCh := make(chan error, sessionCount)

	for i := range sessionCount {

		waitGroup.Go(func() {

			<-start

			sessionID := fmt.Sprintf("session-%d", i)
			session := newTestSession(sessionID)

			if err := repository.Create(
				ctx,
				session,
			); err != nil {
				errCh <- fmt.Errorf(
					"create %q: %w",
					sessionID,
					err,
				)
				return
			}

			loaded, err := repository.GetById(
				ctx,
				sessionID,
			)
			if err != nil {
				errCh <- fmt.Errorf(
					"get %q: %w",
					sessionID,
					err,
				)
				return
			}

			loaded.Score = i

			if err := repository.Update(
				ctx,
				loaded,
			); err != nil {
				errCh <- fmt.Errorf(
					"update %q: %w",
					sessionID,
					err,
				)
			}
		})
	}

	close(start)

	waitGroup.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}

	if t.Failed() {
		return
	}

	for i := range sessionCount {
		sessionID := fmt.Sprintf("session-%d", i)

		loaded, err := repository.GetById(
			ctx,
			sessionID,
		)
		if err != nil {
			t.Fatalf(
				"GetById(%q) error = %v",
				sessionID,
				err,
			)
		}

		if loaded.Score != i {
			t.Errorf(
				"session %q Score = %d, want %d",
				sessionID,
				loaded.Score,
				i,
			)
		}
	}
}

func newTestSession(
	id string,
) *scenario.TrainingSession {
	startedAt := time.Date(
		2026,
		time.August,
		6,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	return &scenario.TrainingSession{
		ID:            id,
		UserID:        42,
		ScenarioID:    1,
		CurrentNodeID: "start",
		Status:        scenario.SessionStatusInProgress,
		Score:         0,
		StartedAt:     startedAt,
		UpdatedAt:     startedAt,
		CompletedAt:   nil,
	}
}

func requireSessionsEqual(
	t *testing.T,
	got *scenario.TrainingSession,
	want *scenario.TrainingSession,
) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"sessions differ:\ngot:  %#v\nwant: %#v",
			got,
			want,
		)
	}
}
