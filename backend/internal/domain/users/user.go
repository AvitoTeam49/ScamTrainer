package usersdomain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrProgressNotFound  = errors.New("user progress not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Score     int       `json:"score" db:"score"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UserProgress struct {
	UserID             int64 `json:"user_id"`
	ScenariosCompleted int   `json:"scenarios_completed"`
	ScamsDetected      int   `json:"scams_detected"`
	FailedAttempts     int   `json:"failed_attempts"`
}

type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Score    int    `json:"score"`
}

type Leaderboard struct {
	TopUsers   []LeaderboardEntry `json:"top_users"`
	TotalUsers int                `json:"total_users"`
}
