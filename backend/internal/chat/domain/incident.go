package domain

import (
	"context"
	"sort"
	"time"
)

type IncidentType string

const (
	IncidentTypeLeftPlatform          IncidentType = "left_platform"
	IncidentTypeDisclosedPersonalData IncidentType = "disclosed_personal_data"
	IncidentTypeFollowedPhishingLink  IncidentType = "followed_phishing_link"
	IncidentTypeInstalledSoftware     IncidentType = "installed_software"
	IncidentTypeDisclosedCardData     IncidentType = "disclosed_card_data"
	IncidentTypeDisclosedCode         IncidentType = "disclosed_code"
	IncidentTypeMadeTransfer          IncidentType = "made_transfer"
)

var incidentWeights = map[IncidentType]int64{
	IncidentTypeLeftPlatform:          10,
	IncidentTypeDisclosedPersonalData: 20,
	IncidentTypeFollowedPhishingLink:  30,
	IncidentTypeInstalledSoftware:     40,
	IncidentTypeDisclosedCardData:     50,
	IncidentTypeDisclosedCode:         60,
	IncidentTypeMadeTransfer:          100,
}

func (t IncidentType) Weight() int64 {
	return incidentWeights[t]
}

func (t IncidentType) Valid() bool {
	_, ok := incidentWeights[t]
	return ok
}

func IncidentTypes() []IncidentType {
	types := make([]IncidentType, 0, len(incidentWeights))
	for incidentType := range incidentWeights {
		types = append(types, incidentType)
	}

	sort.Slice(types, func(i, j int) bool {
		return incidentWeights[types[i]] < incidentWeights[types[j]]
	})

	return types
}

type Incident struct {
	ID        int64
	ChatID    int64
	Type      IncidentType
	Comment   string
	CreatedAt time.Time
}

type IncidentRepository interface {
	ListByChatID(ctx context.Context, chatID int64, cursor Cursor) ([]*Incident, error)
	Create(ctx context.Context, incident *Incident, score int64) error
}
