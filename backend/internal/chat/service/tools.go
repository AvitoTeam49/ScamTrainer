package service

import (
	"encoding/json"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/agent"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/domain"
)

const (
	toolReportIncident = "report_incident"
	toolFinishChat     = "finish_chat"
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

var toolDefinitions = []agent.ToolDefinition{
	{
		Name:        toolReportIncident,
		Description: "Зафиксировать, что собеседник поддался на приём мошенника. Вызывается сразу после реплики, в которой это произошло.",
		Parameters:  reportIncidentSchema(),
	},
	{
		Name:        toolFinishChat,
		Description: "Завершить тренировку и отдать разбор диалога. Вызывается один раз, когда сценарий отыгран или обман раскрыт.",
		Parameters:  finishChatSchema(),
	},
}

func reportIncidentSchema() []byte {
	incidentTypes := domain.IncidentTypes()
	values := make([]string, 0, len(incidentTypes))
	for _, incidentType := range incidentTypes {
		values = append(values, string(incidentType))
	}

	return mustMarshalSchema(jsonSchema{
		Type: "object",
		Properties: map[string]jsonProperty{
			"type": {
				Type:        "string",
				Description: "Тип приёма, на который поддался собеседник",
				Enum:        values,
			},
			"comment": {
				Type:        "string",
				Description: "Короткое объяснение, что именно собеседник сделал неправильно",
			},
		},
		Required:             []string{"type", "comment"},
		AdditionalProperties: false,
	})
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
