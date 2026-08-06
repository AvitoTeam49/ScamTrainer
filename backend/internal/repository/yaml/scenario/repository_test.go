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
		validScenarioYAML("seller_fake_delivery"),
	)

	repository, err := NewYAMLRepository(directory)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	loaded, err := repository.GetByID(
		context.Background(),
		"seller_fake_delivery",
	)
	if err != nil {
		t.Fatalf("get scenario: %v", err)
	}

	if loaded.ID != "seller_fake_delivery" {
		t.Fatalf(
			"unexpected scenario id: got %q",
			loaded.ID,
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

	// В YAML поле ID внутри ноды отсутствует.
	// Его должен заполнить loadScenario из ключа map.
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

	_, err = repository.GetByID(
		context.Background(),
		"missing_scenario",
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
		validScenarioYAML("duplicate_id"),
	)

	writeScenarioFile(
		t,
		directory,
		"second.yaml",
		validScenarioYAML("duplicate_id"),
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
id: broken
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
id: invalid_scenario
title: Invalid scenario
role: seller
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

func validScenarioYAML(id string) string {
	return fmt.Sprintf(`
id: %s
title: Test scenario
role: seller
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
`, id)
}
