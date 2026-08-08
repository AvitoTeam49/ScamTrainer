package scenarioyaml

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
	"gopkg.in/yaml.v3"
)

var (
	_ scenariodomain.ScenarioRepository = (*YAMLRepository)(nil)
	_ scenariodomain.ScenarioCatalog    = (*YAMLRepository)(nil)
)

type YAMLRepository struct {
	scenarios map[int]*scenariodomain.Scenario
}

func NewYAMLRepository(directory string) (*YAMLRepository, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read scenarios directory: %w", err)
	}

	repository := &YAMLRepository{
		scenarios: make(map[int]*scenariodomain.Scenario)}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		extension := filepath.Ext(entry.Name())
		if extension != ".yaml" && extension != ".yml" {
			continue
		}

		path := filepath.Join(directory, entry.Name())

		loaded, err := loadScenario(path)
		if err != nil {
			return nil, err
		}

		_, exists := repository.scenarios[loaded.ID]

		if exists {
			return nil, fmt.Errorf(
				"duplicate scenario id %d",
				loaded.ID,
			)
		}

		repository.scenarios[loaded.ID] = loaded
	}

	return repository, nil
}

func loadScenario(path string) (*scenariodomain.Scenario, error) {
	date, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("read scenario file %q: %w", path, err)
	}

	var loaded scenariodomain.Scenario

	if err := yaml.Unmarshal(date, &loaded); err != nil {
		return nil, fmt.Errorf("unmarshal scenario file %q: %w", path, err)
	}
	for nodeID, node := range loaded.Nodes {
		if node == nil {
			return nil, fmt.Errorf(
				"scenario %d contains empty node %q",
				loaded.ID,
				nodeID,
			)
		}

		node.ID = nodeID
	}

	if err := scenariodomain.Validate(&loaded); err != nil {
		return nil, fmt.Errorf("validate scenario %q: %w", path, err)
	}

	return &loaded, nil
}

func (r *YAMLRepository) GetById(
	_ context.Context,
	scenarioID int,
) (*scenariodomain.Scenario, error) {
	found, exists := r.scenarios[scenarioID]
	if !exists {
		return nil, fmt.Errorf(
			"%w: %d",
			scenariodomain.ErrScenarioNotFound,
			scenarioID,
		)
	}

	return found, nil
}

func scenarioInfoFrom(s *scenariodomain.Scenario) scenariodomain.ScenarioInfo {
	return scenariodomain.ScenarioInfo{
		ID:         s.ID,
		Title:      s.Title,
		Role:       s.Role,
		Difficulty: s.Difficulty,
	}
}

// Метаданные сценариев
func (r *YAMLRepository) List(
	ctx context.Context,
) ([]scenariodomain.ScenarioInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make(
		[]scenariodomain.ScenarioInfo,
		0,
		len(r.scenarios),
	)

	for _, s := range r.scenarios {
		result = append(
			result,
			scenarioInfoFrom(s),
		)
	}

	return result, nil
}

// Метаданные сценариев по уровню сложности
func (r *YAMLRepository) ListByDifficulty(
	ctx context.Context,
	difficulty scenariodomain.Difficulty,
) ([]scenariodomain.ScenarioInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make([]scenariodomain.ScenarioInfo, 0)

	for _, s := range r.scenarios {
		if s.Difficulty != difficulty {
			continue
		}

		result = append(
			result,
			scenarioInfoFrom(s),
		)
	}

	return result, nil
}
