package scenario

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ErrScenarioNotFound = errors.New("scenario not found")

type ScenarioRepository interface {
	GetById(ctx context.Context, id string) (*Scenario, error)
}

type ScenarioCatalog interface {
	List(
		ctx context.Context,
	) ([]ScenarioInfo, error)
}

type YAMLRepository struct {
	scenarios map[string]*Scenario
}

type ScenarioInfo struct {
	ID         string
	Title      string
	Role       Role
	Difficulty Difficulty
}

func NewYAMLRepository(directory string) (*YAMLRepository, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read scenarios directory: %w", err)
	}

	repository := &YAMLRepository{
		scenarios: make(map[string]*Scenario)}

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
				"duplicate scenario id %q",
				loaded.ID,
			)
		}

		repository.scenarios[loaded.ID] = loaded
	}

	return repository, nil
}

func loadScenario(path string) (*Scenario, error) {
	date, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("read scenario file %q: %w", path, err)
	}

	var loaded Scenario

	if err := yaml.Unmarshal(date, &loaded); err != nil {
		return nil, fmt.Errorf("unmarshal scenario file %q: %w", path, err)
	}
	for nodeID, node := range loaded.Nodes {
		if node == nil {
			return nil, fmt.Errorf(
				"scenario %q contains empty node %q",
				loaded.ID,
				nodeID,
			)
		}

		node.ID = nodeID
	}

	if err := Validate(&loaded); err != nil {
		return nil, fmt.Errorf("validate scenario %q: %w", path, err)
	}

	return &loaded, nil
}

func (r *YAMLRepository) GetById(
	_ context.Context,
	scenarioID string,
) (*Scenario, error) {
	found, exists := r.scenarios[scenarioID]
	if !exists {
		return nil, fmt.Errorf(
			"%w: %s",
			ErrScenarioNotFound,
			scenarioID,
		)
	}

	return found, nil
}

func (r *YAMLRepository) List(
	ctx context.Context,
) ([]ScenarioInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make(
		[]ScenarioInfo,
		0,
		len(r.scenarios),
	)

	for _, s := range r.scenarios {
		result = append(result, ScenarioInfo{
			ID:         s.ID,
			Title:      s.Title,
			Role:       s.Role,
			Difficulty: s.Difficulty,
		})
	}

	return result, nil
}
