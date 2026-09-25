package lobby

import (
	"encoding/json"

	"github.com/AlexWendland/go-games-site/internal/domain"
)

const (
	MoveAction = iota
	LobbyAction
	AiAction
)

type GameAction struct {
	ActionType    int
	UserId        domain.PlayerId
	ActionPayload json.RawMessage
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type GameUpdate struct {
	MessageType    string          `json:"type"`
	SequenceNumber int             `json:"sequence_number"`
	Payload        json.RawMessage `json:"payload"`
}

type LobbyActionPayload struct {
	Action   string                      `json:"action"`
	Position domain.PlayerPositionNumber `json:"position"`
}
