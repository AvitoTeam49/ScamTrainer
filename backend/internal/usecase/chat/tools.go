package chatusecase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/agent"
	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
)

const (
	toolApplyTransition = "apply_transition"
	toolFinishChat      = "finish_chat"
)

type jsonProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type jsonSchema struct {
	Type                 string                  `json:"type"`
	Properties           map[string]jsonProperty `json:"properties"`
	Required             []string                `json:"required"`
	AdditionalProperties bool                    `json:"additionalProperties"`
}

// toolDefinitionsFor builds the tool set for one scenario node. The available
// transition ids differ per node, so the schema cannot be a package-level
// constant: it is derived from the graph on every request.
func toolDefinitionsFor(node *scenariodomain.Node) []agent.ToolDefinition {
	return []agent.ToolDefinition{
		{
			Name: toolApplyTransition,
			Description: "Зафиксировать, какой выбор сделал собеседник в текущей точке сценария. " +
				"Вызывается сразу после его реплики, ровно один раз.",
			Parameters: applyTransitionSchema(node),
		},
		{
			Name:        toolFinishChat,
			Description: "Досрочно завершить тренировку и отдать разбор диалога.",
			Parameters:  finishChatSchema(),
		},
	}
}

func applyTransitionSchema(node *scenariodomain.Node) []byte {
	ids := make([]string, 0, len(node.Transitions))
	descriptions := make([]string, 0, len(node.Transitions))

	for _, transition := range node.Transitions {
		ids = append(ids, transition.ID)
		descriptions = append(descriptions, describeTransition(transition))
	}

	return mustMarshalSchema(jsonSchema{
		Type: "object",
		Properties: map[string]jsonProperty{
			"transition_id": {
				Type:        "string",
				Description: strings.Join(descriptions, "\n"),
				Enum:        ids,
			},
		},
		Required:             []string{"transition_id"},
		AdditionalProperties: false,
	})
}

// describeTransition turns a scenario transition into a hint for the model. The
// description and examples are authored in the scenario YAML precisely so the
// model can match a free-form reply onto one of the graph edges.
func describeTransition(transition scenariodomain.Transition) string {
	described := fmt.Sprintf("%s — %s", transition.ID, strings.TrimSpace(transition.Description))

	if len(transition.Examples) > 0 {
		quoted := make([]string, 0, len(transition.Examples))
		for _, example := range transition.Examples {
			quoted = append(quoted, fmt.Sprintf("«%s»", strings.TrimSpace(example)))
		}

		described += " Примеры: " + strings.Join(quoted, ", ")
	}

	return described
}

func finishChatSchema() []byte {
	return mustMarshalSchema(jsonSchema{
		Type: "object",
		Properties: map[string]jsonProperty{
			"resume": {
				Type:        "string",
				Description: "Разбор диалога: на что собеседник повёлся, что сделал правильно, на что смотреть в следующий раз",
			},
		},
		Required:             []string{"resume"},
		AdditionalProperties: false,
	})
}

func mustMarshalSchema(schema jsonSchema) []byte {
	encoded, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}

	return encoded
}
