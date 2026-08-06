package training

import (
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/scenario"
)

type TrainingService struct {
	scenarios scenario.ScenarioRepository
	sessions  SessionRepository
	engine    *scenario.Engine
	ids       IDGenerator
}

func NewService(
	scenarios scenario.ScenarioRepository,
	sessions SessionRepository,
	engine *scenario.Engine,
	ids IDGenerator,
) *TrainingService {
	return &TrainingService{
		scenarios: scenarios,
		sessions:  sessions,
		engine:    engine,
		ids:       ids,
	}
}
