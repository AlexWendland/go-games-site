package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/AlexWendland/go-games-site/internal/domain"
)

const (
	ChannelSize = 20
)

type Game interface {
	MakeMove(playerPosition domain.PlayerPositionNumber, moveData json.RawMessage) error
	GetPublicState() json.RawMessage
	GetPlayerState(playerPosition domain.PlayerPositionNumber) json.RawMessage
	GetSaveState() json.RawMessage
	GetGameStatus() string
	SetGameState(gameState json.RawMessage)
	GetMaxPlayers() domain.PlayerPositionNumber
}

type GameDbService interface {
	SaveGameEvent(ctx context.Context, gameId string, position domain.PlayerPositionNumber, sequenceNumber int, payload json.RawMessage, createdAt time.Time) error
	SaveGameState(ctx context.Context, gameId string, updateTime time.Time, newState json.RawMessage, gameStatus string) error
	MarkPlayerJoined(ctx context.Context, gameId string, positionToAdd domain.PlayerPosition, joinTime time.Time) error
	MarkPlayerLeft(ctx context.Context, gameId string, positionToRemove domain.PlayerPosition, leaveTime time.Time) error
	GetCurrentPlayerPositions(ctx context.Context, gameId string) ([]domain.PlayerPosition, error)
	// TODO: Define some type to represent an AI player so we can start them.
	GetCurrentAiPlayers(ctx context.Context, gameId string) []string
	GetCurrentEventCount(ctx context.Context, gameId string) (int, error)
	WithTx(ctx context.Context, fn func(GameDbService) error) error
}

type Subscriber struct {
	player  domain.PlayerId
	channel chan<- json.RawMessage
}

type Lobby struct {
	game             Game
	gameId           string
	inputChannel     chan GameAction
	dbService        GameDbService
	subscribers      map[int]Subscriber
	positionToPlayer map[domain.PlayerPositionNumber]domain.PlayerId
	eventCount       int
	subCount         int
	cancel           func()
	subMutex         sync.Mutex
}

func CreateLobby(ctx context.Context, game Game, gameId string, db GameDbService) (*Lobby, error) {
	// Populate player positions based on the database
	positions, err := db.GetCurrentPlayerPositions(ctx, gameId)
	if err != nil {
		return nil, domain.ErrDatabase
	}
	positionToPlayer, err := createPlayerPositionMap(positions)
	if err != nil {
		slog.Error("inconsistent player positions detected", "gameId", gameId, "positions", positions)
		return nil, domain.ErrInternal
	}
	subscribers := make(map[int]Subscriber)
	var subCount = 0
	// TODO: Add startup for AI players here if it is the right place

	currentEventCount, err := db.GetCurrentEventCount(ctx, gameId)
	if err != nil {
		return nil, domain.ErrDatabase
	}

	return &Lobby{game: game, gameId: gameId, inputChannel: make(chan GameAction, ChannelSize), dbService: db, subscribers: subscribers, positionToPlayer: positionToPlayer, eventCount: currentEventCount, subCount: subCount}, nil
}

func (l *Lobby) GetInputChannel() chan<- GameAction {
	return l.inputChannel
}

// Start up the server
func (l *Lobby) Serve(ctx context.Context, onShutdown func()) {
	gameCtx, cancel := context.WithCancel(ctx)
	l.cancel = cancel
	defer onShutdown()
	for {
		select {
		case <-gameCtx.Done():
			return
		case message := <-l.inputChannel:
			switch message.ActionType {
			case MoveAction:
				l.handleMoveAction(gameCtx, message.UserId, message.ActionPayload)
			case LobbyAction:
				l.handleLobbyAction(gameCtx, message.UserId, message.ActionPayload)
			case AiAction:
				l.handleAiAction(gameCtx, message.UserId, message.ActionPayload)
			default:
				slog.Error("Invalid action", "gameId", l.gameId, "action", message)
			}
		}
	}
}

func (l *Lobby) AddSubscriber(playerId domain.PlayerId) (<-chan json.RawMessage, int) {
	l.subMutex.Lock()
	defer l.subMutex.Unlock()

	id := l.subCount
	// Use buffered channel in case of slow reader
	subscriberChannel := make(chan json.RawMessage, ChannelSize)
	l.subscribers[id] = Subscriber{
		player:  playerId,
		channel: subscriberChannel,
	}
	l.subCount++
	return subscriberChannel, id
}

func (l *Lobby) RemoveSubscriber(subId int) {
	l.subMutex.Lock()
	defer l.subMutex.Unlock()

	subscriber, ok := l.subscribers[subId]
	if !ok {
		return
	}
	close(subscriber.channel)
	delete(l.subscribers, subId)

	// Check if there are no human subscribers if so close the context.
	if l.cancel == nil {
		return
	}

	for _, subscriber := range l.subscribers {
		if !subscriber.player.IsAi {
			return
		}
	}
	l.cancel()
}

func (l *Lobby) getPlayerPosition(targetPlayer domain.PlayerId) (domain.PlayerPositionNumber, bool) {
	for position, playerId := range l.positionToPlayer {
		if targetPlayer == playerId {
			return position, true
		}
	}
	var zero domain.PlayerPositionNumber
	return zero, false
}

func createPlayerPositionMap(positions []domain.PlayerPosition) (map[domain.PlayerPositionNumber]domain.PlayerId, error) {
	output := make(map[domain.PlayerPositionNumber]domain.PlayerId)
	seen := make(map[domain.PlayerId]bool)
	for _, playerPosition := range positions {

		// If position already exists exit
		if _, ok := output[playerPosition.Position]; ok {
			return nil, domain.ErrInternal
		}
		// If player already assigned exit
		if seen[playerPosition.Player] {
			return nil, domain.ErrInternal
		}
		seen[playerPosition.Player] = true
		output[playerPosition.Position] = playerPosition.Player
	}
	return output, nil
}

func (l *Lobby) handleMoveAction(ctx context.Context, userId domain.PlayerId, payload json.RawMessage) {
	playerPosition, ok := l.getPlayerPosition(userId)
	if !ok {
		slog.Info("user made move without being in game", "gameId", l.gameId, "userId", userId)
		l.sendErrorToUser(userId, l.eventCount, "can not make move without being in the game")
		return
	}

	// Checkpoint to roll back if saving to the DB fails
	checkpoint := l.game.GetSaveState()

	err := l.game.MakeMove(playerPosition, payload)
	if err != nil {
		slog.Info("user made invalid move", "gameId", l.gameId, "userId", userId, "payload", payload)
		l.sendErrorToUser(userId, l.eventCount, "tried to make invalid move")
		return
	}

	newState := l.game.GetSaveState()
	gameStatus := l.game.GetGameStatus()
	timeOfChange := time.Now()
	sequenceNumber := l.eventCount

	err = l.dbService.WithTx(ctx, func(db GameDbService) error {
		err := db.SaveGameState(ctx, l.gameId, timeOfChange, newState, gameStatus)
		if err != nil {
			return err
		}
		return db.SaveGameEvent(ctx, l.gameId, playerPosition, l.eventCount, payload, timeOfChange)
	})
	if err != nil {
		// Perform roll back
		l.game.SetGameState(checkpoint)
		l.sendErrorToUser(userId, l.eventCount, "internal server error")
		return
	}
	l.eventCount++

	l.broadcast(sequenceNumber, domain.StateUpdateType, func(player domain.PlayerId) json.RawMessage {
		position, ok := l.getPlayerPosition(player)
		if !ok {
			return l.game.GetPublicState()
		}
		return l.game.GetPlayerState(position)
	})
}

func (l *Lobby) handleJoinLobbyAction(ctx context.Context, userId domain.PlayerId, position domain.PlayerPositionNumber) error {
	if _, ok := l.positionToPlayer[position]; ok {
		l.sendErrorToUser(userId, l.eventCount, fmt.Sprintf("position %d already taken", position))
		return domain.ErrUserError
	}
	if position >= l.game.GetMaxPlayers() {
		l.sendErrorToUser(userId, l.eventCount, fmt.Sprintf("position %d exceeds max players %d", position, l.game.GetMaxPlayers()))
		return domain.ErrUserError
	}
	actionTime := time.Now()
	oldPosition, wasInGame := l.getPlayerPosition(userId)
	err := l.dbService.WithTx(ctx, func(tx GameDbService) error {
		if wasInGame {
			err := tx.MarkPlayerLeft(ctx, l.gameId, domain.PlayerPosition{Position: oldPosition, Player: userId}, actionTime)
			if err != nil {
				return err
			}
		}
		return tx.MarkPlayerJoined(ctx, l.gameId, domain.PlayerPosition{Position: position, Player: userId}, actionTime)
	})
	if err != nil {
		l.sendErrorToUser(userId, l.eventCount, "database error please try again")
		slog.Warn("database error to join lobby", "gameId", l.gameId, "position", position, "player", userId)
		return err
	}
	if wasInGame {
		delete(l.positionToPlayer, oldPosition)
		slog.Debug("player left position", "gameId", l.gameId, "position", oldPosition, "player", userId)
	}
	l.positionToPlayer[position] = userId
	slog.Debug("player joined position", "gameId", l.gameId, "position", position, "player", userId)
	return nil
}

func (l *Lobby) handleLeaveLobbyAction(ctx context.Context, userId domain.PlayerId, position domain.PlayerPositionNumber) error {
	playerCurrentPosition, isInGame := l.getPlayerPosition(userId)
	if !isInGame || playerCurrentPosition != position {
		l.sendErrorToUser(userId, l.eventCount, fmt.Sprintf("player is not in position %d", position))
		return domain.ErrUserError
	}
	err := l.dbService.MarkPlayerLeft(ctx, l.gameId, domain.PlayerPosition{Position: position, Player: userId}, time.Now())
	if err != nil {
		slog.Warn("database error to leave lobby", "gameId", l.gameId, "position", position, "player", userId)
		return err
	}
	delete(l.positionToPlayer, position)
	slog.Debug("player left position", "gameId", l.gameId, "position", position, "player", userId)
	return nil
}

// Structure of the lobby action.
//
//	{
//		"action" : "join/leave",
//		"position" : 0/1/...
//	}
func (l *Lobby) handleLobbyAction(ctx context.Context, userId domain.PlayerId, payload json.RawMessage) {
	var action LobbyActionPayload
	if err := json.Unmarshal(payload, &action); err != nil {
		l.sendErrorToUser(userId, l.eventCount, "invalid lobby action payload")
		return
	}
	var err error
	switch action.Action {
	case "join":
		err = l.handleJoinLobbyAction(ctx, userId, action.Position)
	case "leave":
		err = l.handleLeaveLobbyAction(ctx, userId, action.Position)
	default:
		l.sendErrorToUser(userId, l.eventCount, "invalid lobby action")
		err = domain.ErrUserError
	}
	if err != nil {
		return
	}
	// Create lobby payload
	var lobbyPlayers []domain.PlayerPosition
	for position, player := range l.positionToPlayer {
		lobbyPlayers = append(lobbyPlayers, domain.PlayerPosition{Position: position, Player: player})
	}
	lobbyPayload, err := json.Marshal(LobbyUpdatePayload{lobbyPlayers})
	if err != nil {
		slog.Error("failed to marshal lobby messages to send to users", "gameId", l.gameId)
		return
	}
	l.broadcast(l.eventCount, domain.LobbyUpdateType, func(player domain.PlayerId) json.RawMessage {
		return lobbyPayload
	})
	l.eventCount++
}

func (l *Lobby) handleAiAction(context context.Context, userId domain.PlayerId, payload json.RawMessage) {
	// TODO: Implement this if this is the right place for it.
}

func (l *Lobby) broadcast(sequenceNumber int, messageType string, payloadFor func(domain.PlayerId) json.RawMessage) {
	deadIds := make([]int, 0)

	l.subMutex.Lock()
	for id, subscriber := range l.subscribers {
		message := GameUpdate{
			MessageType:    messageType,
			SequenceNumber: sequenceNumber,
			Payload:        payloadFor(subscriber.player),
		}
		encodedMessage, err := json.Marshal(message)
		if err != nil {
			slog.Warn("failed to marshal message to send to user", "gameId", l.gameId, "userId", subscriber.player)
			errPayload, err := json.Marshal(ErrorPayload{Message: "missing update"})
			if err != nil {
				continue
			}
			errMessage := GameUpdate{
				MessageType:    domain.ErrorUpdateType,
				SequenceNumber: sequenceNumber,
				Payload:        errPayload,
			}
			encodedMessage, err = json.Marshal(errMessage)
			if err != nil {
				continue
			}
		}
		select {
		case subscriber.channel <- encodedMessage:
		default:
			deadIds = append(deadIds, id)
		}
	}
	l.subMutex.Unlock()

	for _, id := range deadIds {
		l.RemoveSubscriber(id)
	}
}

func (l *Lobby) sendErrorToUser(userId domain.PlayerId, sequenceNumber int, errorMessage string) {
	deadIds := make([]int, 0)

	l.subMutex.Lock()
	for id, subscriber := range l.subscribers {
		if userId != subscriber.player {
			continue
		}
		encodedPayload, err := json.Marshal(ErrorPayload{Message: errorMessage})
		if err != nil {
			slog.Warn("failed to marshal error payload", "gameId", l.gameId, "userId", subscriber.player)
			continue
		}
		message := GameUpdate{
			MessageType:    domain.ErrorUpdateType,
			SequenceNumber: sequenceNumber,
			Payload:        encodedPayload,
		}
		encodedMessage, err := json.Marshal(message)
		if err != nil {
			slog.Warn("failed to marshal message to send to user", "gameId", l.gameId, "userId", subscriber.player)
			continue
		}
		select {
		case subscriber.channel <- encodedMessage:
		default:
			deadIds = append(deadIds, id)
		}
	}
	l.subMutex.Unlock()

	for _, id := range deadIds {
		l.RemoveSubscriber(id)
	}
}
