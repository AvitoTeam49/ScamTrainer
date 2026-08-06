package scenariodomain

import (
	"strings"
	"testing"
)

func TestValidate_ValidDAG(t *testing.T) {
	scenario := newValidScenario()

	if err := Validate(scenario); err != nil {
		t.Fatalf("expected valid scenario, got error: %v", err)
	}
}

func TestValidate_ValidDAGWithMergedBranches(t *testing.T) {
	scenario := newValidScenario()

	if err := Validate(scenario); err != nil {
		t.Fatalf(
			"expected DAG with merged branches to be valid, got: %v",
			err,
		)
	}
}

func TestValidate_NilScenario(t *testing.T) {
	err := Validate(nil)
	if err == nil {
		t.Fatal("expected validation error for nil scenario")
	}
}

func TestValidate_InvalidScenarios(t *testing.T) {
	tests := []struct {
		name          string
		prepare       func(*Scenario)
		wantErrorPart string
	}{
		{
			name: "empty scenario id",
			prepare: func(s *Scenario) {
				s.ID = ""
			},
			wantErrorPart: "id",
		},
		{
			name: "empty title",
			prepare: func(s *Scenario) {
				s.Title = ""
			},
			wantErrorPart: "title",
		},
		{
			name: "unsupported role",
			prepare: func(s *Scenario) {
				s.Role = Role("administrator")
			},
			wantErrorPart: "role",
		},
		{
			name: "no nodes",
			prepare: func(s *Scenario) {
				s.Nodes = nil
			},
			wantErrorPart: "node",
		},
		{
			name: "missing start node",
			prepare: func(s *Scenario) {
				s.StartNodeID = "missing_start"
			},
			wantErrorPart: "start",
		},
		{
			name: "nil node",
			prepare: func(s *Scenario) {
				s.Nodes["left"] = nil
			},
			wantErrorPart: "nil",
		},
		{
			name: "unsupported node type",
			prepare: func(s *Scenario) {
				s.Nodes["left"].Type = NodeType("unknown")
			},
			wantErrorPart: "type",
		},
		{
			name: "decision without transitions",
			prepare: func(s *Scenario) {
				s.Nodes["left"].Transitions = nil
			},
			wantErrorPart: "transition",
		},
		{
			name: "ending with transition",
			prepare: func(s *Scenario) {
				s.Nodes["safe_ending"].Transitions = []Transition{
					{
						ID:          "invalid",
						Description: "Invalid transition",
						ToNodeID:    "start",
						Feedback:    "Invalid",
					},
				}
			},
			wantErrorPart: "ending",
		},
		{
			name: "transition without id",
			prepare: func(s *Scenario) {
				s.Nodes["start"].Transitions[0].ID = ""
			},
			wantErrorPart: "id",
		},
		{
			name: "duplicate transition id",
			prepare: func(s *Scenario) {
				s.Nodes["start"].Transitions[1].ID =
					s.Nodes["start"].Transitions[0].ID
			},
			wantErrorPart: "duplicate",
		},
		{
			name: "transition with empty target",
			prepare: func(s *Scenario) {
				s.Nodes["start"].Transitions[0].ToNodeID = ""
			},
			wantErrorPart: "target",
		},
		{
			name: "transition to missing node",
			prepare: func(s *Scenario) {
				s.Nodes["start"].Transitions[0].ToNodeID =
					"missing_node"
			},
			wantErrorPart: "missing_node",
		},
		{
			name: "unreachable node",
			prepare: func(s *Scenario) {
				s.Nodes["unreachable"] = &Node{
					ID:      "unreachable",
					Type:    NodeTypeEnding,
					Title:   "Unreachable ending",
					Outcome: "safe",
					Message: Message{
						Author: "system",
						Text:   "This node cannot be reached",
					},
				}
			},
			wantErrorPart: "unreachable",
		},
		{
			name: "cycle",
			prepare: func(s *Scenario) {
				ending := s.Nodes["safe_ending"]

				ending.Type = NodeTypeDecision
				ending.Outcome = ""
				ending.Transitions = []Transition{
					{
						ID:          "return_to_start",
						Description: "Return to start",
						ToNodeID:    "start",
						Feedback:    "Cycle",
					},
				}
			},
			wantErrorPart: "cycle",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scenario := newValidScenario()
			test.prepare(scenario)

			err := Validate(scenario)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			if test.wantErrorPart != "" &&
				!strings.Contains(
					strings.ToLower(err.Error()),
					strings.ToLower(test.wantErrorPart),
				) {
				t.Fatalf(
					"expected error containing %q, got: %v",
					test.wantErrorPart,
					err,
				)
			}
		})
	}
}

func newValidScenario() *Scenario {
	return &Scenario{
		ID:          "test_scenario",
		Title:       "Test scenario",
		Role:        RoleSeller,
		StartNodeID: "start",
		LLM: ScenarioLLM{
			CharacterPrompt: "Act as another platform user",
		},
		Nodes: map[string]*Node{
			"start": {
				ID:   "start",
				Type: NodeTypeDecision,
				Message: Message{
					Author: "scammer",
					Text:   "Choose what to do",
				},
				LLM: NodeLLM{
					ReplyInstruction: "Wait for the user response",
				},
				Transitions: []Transition{
					{
						ID:          "go_left",
						Description: "Choose the left branch",
						Examples: []string{
							"I want to inspect the offer",
						},
						ToNodeID:   "left",
						ScoreDelta: 0,
						Feedback:   "The scenario continues",
					},
					{
						ID:          "go_right",
						Description: "Choose the right branch",
						Examples: []string{
							"I refuse the external link",
						},
						ToNodeID:   "right",
						ScoreDelta: 10,
						Feedback:   "The scenario continues safely",
					},
				},
			},
			"left": {
				ID:   "left",
				Type: NodeTypeDecision,
				Message: Message{
					Author: "scammer",
					Text:   "Please continue",
				},
				LLM: NodeLLM{
					ReplyInstruction: "Continue the dialogue",
				},
				Transitions: []Transition{
					{
						ID:          "finish_from_left",
						Description: "Finish the left branch",
						Examples: []string{
							"I will stay on the platform",
						},
						ToNodeID:   "safe_ending",
						ScoreDelta: 10,
						Feedback:   "Safe decision",
					},
				},
			},
			"right": {
				ID:   "right",
				Type: NodeTypeDecision,
				Message: Message{
					Author: "scammer",
					Text:   "Are you sure?",
				},
				LLM: NodeLLM{
					ReplyInstruction: "Ask the user to confirm",
				},
				Transitions: []Transition{
					{
						ID:          "finish_from_right",
						Description: "Finish the right branch",
						Examples: []string{
							"Yes, I refuse",
						},
						ToNodeID:   "safe_ending",
						ScoreDelta: 10,
						Feedback:   "Safe decision",
					},
				},
			},
			"safe_ending": {
				ID:      "safe_ending",
				Type:    NodeTypeEnding,
				Title:   "Safe ending",
				Outcome: "safe",
				Message: Message{
					Author: "system",
					Text:   "The scenario is complete",
				},
			},
		},
	}
}
