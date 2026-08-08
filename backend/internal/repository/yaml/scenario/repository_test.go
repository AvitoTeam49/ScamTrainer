package scenarioyaml

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
)

func TestNewYAMLRepository_LoadsScenario(t *testing.T) {
	directory := t.TempDir()

	writeScenarioFile(
		t,
		directory,
		"seller.yaml",
		validScenarioYAML(1),
	)

	repository, err := NewYAMLRepository(directory)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	loaded, err := repository.GetById(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("get scenario: %v", err)
	}

	if loaded.ID != 1 {
		t.Fatalf(
			"unexpected scenario id: got %d",
			loaded.ID,
		)
	}

	if loaded.Difficulty != scenariodomain.DifficultyEasy {
		t.Fatalf(
			"unexpected difficulty: got %d",
			loaded.Difficulty,
		)
	}

	if loaded.StartNodeID != "start" {
		t.Fatalf(
			"unexpected start node id: got %q",
			loaded.StartNodeID,
		)
	}

	startNode := loaded.Nodes["start"]
	if startNode == nil {
		t.Fatal("start node is nil")
	}

	if startNode.ID != "start" {
		t.Fatalf(
			"unexpected node id: got %q",
			startNode.ID,
		)
	}
}

func TestYAMLRepository_GetById_NotFound(t *testing.T) {
	directory := t.TempDir()

	repository, err := NewYAMLRepository(directory)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	_, err = repository.GetById(
		context.Background(),
		100,
	)
	if err == nil {
		t.Fatal("expected scenario not found error")
	}

	if !errors.Is(err, scenariodomain.ErrScenarioNotFound) {
		t.Fatalf(
			"expected ErrScenarioNotFound, got: %v",
			err,
		)
	}
}

func TestNewYAMLRepository_DuplicateScenarioID(t *testing.T) {
	directory := t.TempDir()

	writeScenarioFile(
		t,
		directory,
		"first.yaml",
		validScenarioYAML(1),
	)

	writeScenarioFile(
		t,
		directory,
		"second.yaml",
		validScenarioYAML(1),
	)

	_, err := NewYAMLRepository(directory)
	if err == nil {
		t.Fatal("expected duplicate scenario id error")
	}

	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf(
			"expected duplicate error, got: %v",
			err,
		)
	}
}

func TestNewYAMLRepository_MalformedYAML(t *testing.T) {
	directory := t.TempDir()

	writeScenarioFile(
		t,
		directory,
		"broken.yaml",
		`
id: 3
title: Broken
nodes:
  start:
    type: [
`,
	)

	_, err := NewYAMLRepository(directory)
	if err == nil {
		t.Fatal("expected YAML parsing error")
	}
}

func TestNewYAMLRepository_InvalidScenario(t *testing.T) {
	directory := t.TempDir()

	writeScenarioFile(
		t,
		directory,
		"invalid.yaml",
		`
id: 2
title: Invalid scenario
role: seller
difficulty: 0
start_node_id: start

nodes:
  start:
    type: decision
    transitions:
      - id: missing_target
        description: Invalid transition
        to_node_id: unknown_node
        score_delta: 0
`,
	)

	_, err := NewYAMLRepository(directory)
	if err == nil {
		t.Fatal("expected scenario validation error")
	}

	if !strings.Contains(err.Error(), "unknown_node") {
		t.Fatalf(
			"expected missing target error, got: %v",
			err,
		)
	}
}

func TestNewYAMLRepository_MissingDirectory(t *testing.T) {
	directory := filepath.Join(
		t.TempDir(),
		"does-not-exist",
	)

	_, err := NewYAMLRepository(directory)
	if err == nil {
		t.Fatal("expected directory reading error")
	}
}

func writeScenarioFile(
	t *testing.T,
	directory string,
	filename string,
	content string,
) {
	t.Helper()

	path := filepath.Join(directory, filename)

	if err := os.WriteFile(
		path,
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatalf(
			"write scenario file %q: %v",
			path,
			err,
		)
	}
}

func validScenarioYAML(id int) string {
	return validScenarioYAMLWithDifficulty(
		id,
		scenariodomain.DifficultyEasy,
	)
}

func validScenarioYAMLWithDifficulty(
	id int,
	difficulty scenariodomain.Difficulty,
) string {
	return fmt.Sprintf(`
id: %d
title: Test scenario
role: seller
difficulty: %d
start_node_id: start

llm:
  character_prompt: Test character prompt

nodes:
  start:
    type: decision

    message:
      author: scammer
      text: Open this external link
    llm:
      reply_instruction: Persuade the user

    transitions:
      - id: stay_safe
        description: Stay on the platform
        examples:
          - I will not open the link
        to_node_id: safe_ending
        score_delta: 20
        feedback: Safe action
        risk_tags:
          - external_link
      - id: open_link
        description: Open the external link
        examples:
          - I will open it
        to_node_id: unsafe_ending
        score_delta: -20
        feedback: Unsafe action
        risk_tags:
          - phishing

  safe_ending:
    type: ending
    title: Safe ending
    outcome: safe

    message:
      author: system
      text: The user stayed safe

  unsafe_ending:
    type: ending
    title: Unsafe ending
    outcome: unsafe
    message:
      author: system
      text: The user opened a phishing link
`, id, difficulty)
}

func TestYAMLRepository_List(t *testing.T) {
	directory := t.TempDir()

	writeScenarioFile(
		t,
		directory,
		"first.yaml",
		validScenarioYAMLWithDifficulty(
			1,
			scenariodomain.DifficultyEasy,
		),
	)

	writeScenarioFile(
		t,
		directory,
		"second.yaml",
		validScenarioYAMLWithDifficulty(
			2,
			scenariodomain.DifficultyMedium,
		),
	)

	repository, err := NewYAMLRepository(directory)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	scenarios, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("list scenarios: %v", err)
	}

	if len(scenarios) != 2 {
		t.Fatalf(
			"unexpected scenarios count: got %d, want 2",
			len(scenarios),
		)
	}

	byID := make(map[int]scenariodomain.ScenarioInfo, len(scenarios))
	for _, scenario := range scenarios {
		byID[scenario.ID] = scenario
	}

	first, exists := byID[1]
	if !exists {
		t.Fatal("scenario 1 not found")
	}

	if first.Title != "Test scenario" {
		t.Errorf(
			"unexpected title: got %q",
			first.Title,
		)
	}

	if first.Role != scenariodomain.RoleSeller {
		t.Errorf(
			"unexpected role: got %q",
			first.Role,
		)
	}

	if first.Difficulty != scenariodomain.DifficultyEasy {
		t.Errorf(
			"unexpected difficulty: got %d",
			first.Difficulty,
		)
	}

	second, exists := byID[2]
	if !exists {
		t.Fatal("scenario 2 not found")
	}

	if second.Difficulty != scenariodomain.DifficultyMedium {
		t.Errorf(
			"unexpected difficulty: got %d",
			second.Difficulty,
		)
	}
}

func TestYAMLRepository_ListByDifficulty(t *testing.T) {
	directory := t.TempDir()

	writeScenarioFile(
		t,
		directory,
		"easy.yaml",
		validScenarioYAMLWithDifficulty(
			1,
			scenariodomain.DifficultyEasy,
		),
	)

	writeScenarioFile(
		t,
		directory,
		"medium-first.yaml",
		validScenarioYAMLWithDifficulty(
			2,
			scenariodomain.DifficultyMedium,
		),
	)

	writeScenarioFile(
		t,
		directory,
		"medium-second.yaml",
		validScenarioYAMLWithDifficulty(
			3,
			scenariodomain.DifficultyMedium,
		),
	)

	repository, err := NewYAMLRepository(directory)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	scenarios, err := repository.ListByDifficulty(
		context.Background(),
		scenariodomain.DifficultyMedium,
	)
	if err != nil {
		t.Fatalf("list scenarios by difficulty: %v", err)
	}

	if len(scenarios) != 2 {
		t.Fatalf(
			"unexpected scenarios count: got %d, want 2",
			len(scenarios),
		)
	}

	byID := make(map[int]scenariodomain.ScenarioInfo, len(scenarios))
	for _, scenario := range scenarios {
		byID[scenario.ID] = scenario
	}

	if _, exists := byID[1]; exists {
		t.Fatal("easy scenario must not be returned")
	}

	for _, id := range []int{2, 3} {
		scenario, exists := byID[id]
		if !exists {
			t.Fatalf(
				"scenario %d not found",
				id,
			)
		}

		if scenario.Difficulty != scenariodomain.DifficultyMedium {
			t.Errorf(
				"scenario %d has unexpected difficulty: got %d",
				id,
				scenario.Difficulty,
			)
		}
	}
}

func TestYAMLRepository_ListByDifficulty_NoMatches(
	t *testing.T,
) {
	directory := t.TempDir()

	writeScenarioFile(
		t,
		directory,
		"easy.yaml",
		validScenarioYAMLWithDifficulty(
			1,
			scenariodomain.DifficultyEasy,
		),
	)

	repository, err := NewYAMLRepository(directory)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	scenarios, err := repository.ListByDifficulty(
		context.Background(),
		scenariodomain.DifficultyHard,
	)
	if err != nil {
		t.Fatalf("list scenarios by difficulty: %v", err)
	}

	if len(scenarios) != 0 {
		t.Fatalf(
			"expected no scenarios, got %d",
			len(scenarios),
		)
	}
}
