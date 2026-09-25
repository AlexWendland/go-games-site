package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/AlexWendland/go-games-site/internal/domain"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	Player1   = domain.PlayerId{UserId: "Player1", IsAi: false}
	Player2   = domain.PlayerId{UserId: "Player2", IsAi: false}
	Player1Ai = domain.PlayerId{UserId: "Player1", IsAi: true}
)

type GameStub struct {
	makeMove       func(playerPosition domain.PlayerPositionNumber, moveData json.RawMessage) error
	getPublicState func() json.RawMessage
	getPlayerState func(playerPosition domain.PlayerPositionNumber) json.RawMessage
	getSaveState   func() json.RawMessage
	getGameStatus  func() string
	setGameState   func(gameState json.RawMessage)
}

func (g GameStub) MakeMove(playerPosition domain.PlayerPositionNumber, moveData json.RawMessage) error {
	return g.makeMove(playerPosition, moveData)
}

func (g GameStub) GetPublicState() json.RawMessage {
	return g.getPublicState()
}

func (g GameStub) GetPlayerState(playerPosition domain.PlayerPositionNumber) json.RawMessage {
	return g.getPlayerState(playerPosition)
}

func (g GameStub) GetSaveState() json.RawMessage {
	return g.getSaveState()
}

func (g GameStub) GetGameStatus() string {
	return g.getGameStatus()
}

func (g GameStub) SetGameState(gameState json.RawMessage) {
	g.setGameState(gameState)
}

type GameDbServiceStub struct {
	saveGameEvent             func(ctx context.Context, gameId string, userId domain.PlayerId, sequenceNumber int, eventType string, payload json.RawMessage, createdAt time.Time) error
	saveGameState             func(ctx context.Context, gameId string, updateTime time.Time, newState json.RawMessage, gameStatus string) error
	markPlayerJoined          func(ctx context.Context, gameId string, positionToAdd domain.PlayerPosition, joinTime time.Time) error
	markPlayerLeft            func(ctx context.Context, gameId string, positionToRemove domain.PlayerPosition, leaveTime time.Time) error
	getCurrentPlayerPositions func(ctx context.Context, gameId string) ([]domain.PlayerPosition, error)
	getCurrentAiPlayers       func(ctx context.Context, gameId string) []string
	getCurrentEventCount      func(ctx context.Context, gameId string) (int, error)
	withTx                    func(ctx context.Context, fn func(GameDbService) error) error
}

func (s GameDbServiceStub) SaveGameEvent(ctx context.Context, gameId string, userId domain.PlayerId, sequenceNumber int, eventType string, payload json.RawMessage, createdAt time.Time) error {
	return s.saveGameEvent(ctx, gameId, userId, sequenceNumber, eventType, payload, createdAt)
}

func (s GameDbServiceStub) SaveGameState(ctx context.Context, gameId string, updateTime time.Time, newState json.RawMessage, gameStatus string) error {
	return s.saveGameState(ctx, gameId, updateTime, newState, gameStatus)
}

func (s GameDbServiceStub) MarkPlayerJoined(ctx context.Context, gameId string, positionToAdd domain.PlayerPosition, joinTime time.Time) error {
	return s.markPlayerJoined(ctx, gameId, positionToAdd, joinTime)
}

func (s GameDbServiceStub) MarkPlayeLeft(ctx context.Context, gameId string, positionToRemove domain.PlayerPosition, leaveTime time.Time) error {
	return s.markPlayerLeft(ctx, gameId, positionToRemove, leaveTime)
}

func (s GameDbServiceStub) GetCurrentPlayerPositions(ctx context.Context, gameId string) ([]domain.PlayerPosition, error) {
	return s.getCurrentPlayerPositions(ctx, gameId)
}

func (s GameDbServiceStub) GetCurrentAiPlayers(ctx context.Context, gameId string) []string {
	return s.getCurrentAiPlayers(ctx, gameId)
}

func (s GameDbServiceStub) GetCurrentEventCount(ctx context.Context, gameId string) (int, error) {
	return s.getCurrentEventCount(ctx, gameId)
}

func (s GameDbServiceStub) WithTx(ctx context.Context, fn func(GameDbService) error) error {
	return s.withTx(ctx, fn)
}

func Test_createPlayerPositionMap(t *testing.T) {
	tests := []struct {
		name        string
		positions   []domain.PlayerPosition
		expectedErr error
		expectedMap map[domain.PlayerPositionNumber]domain.PlayerId
	}{
		{"empty start", []domain.PlayerPosition{}, nil, map[domain.PlayerPositionNumber]domain.PlayerId{}},
		{"conflicting positions", []domain.PlayerPosition{{Position: 0, Player: Player1}, {Position: 0, Player: Player2}}, domain.ErrInternal, nil},
		{"conflicting positions ai", []domain.PlayerPosition{{Position: 0, Player: Player1}, {Position: 0, Player: Player1Ai}}, domain.ErrInternal, nil},
		{"player in two positions", []domain.PlayerPosition{{Position: 0, Player: Player1}, {Position: 1, Player: Player1}}, domain.ErrInternal, nil},
		{"working", []domain.PlayerPosition{{Position: 0, Player: Player1}, {Position: 1, Player: Player2}, {Position: 3, Player: Player1Ai}}, nil, map[domain.PlayerPositionNumber]domain.PlayerId{0: Player1, 1: Player2, 3: Player1Ai}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualMap, actualErr := createPlayerPositionMap(tt.positions)
			if !errors.Is(actualErr, tt.expectedErr) {
				t.Errorf("got err %v, want %v", actualErr, tt.expectedErr)
			}
			if diff := cmp.Diff(tt.expectedMap, actualMap); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestAddSubscriber(t *testing.T) {
	t.Run("stores subscriber and returns non-nil channel", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		ch, id := l.AddSubscriber(Player1)
		if id != 0 {
			t.Errorf("sub id = %d, want 0", id)
		}
		if ch == nil {
			t.Error("expected non-nil channel")
		}
		if sub, ok := l.subscribers[id]; !ok {
			t.Error("subscriber not stored")
		} else if sub.player != Player1 {
			t.Errorf("stored player = %v, want %v", sub.player, Player1)
		}
	})

	t.Run("ids increment across multiple subscribers", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		_, id1 := l.AddSubscriber(Player1)
		_, id2 := l.AddSubscriber(Player2)
		if id1 != 0 {
			t.Errorf("first sub id = %d, want 0", id1)
		}
		if id2 != 1 {
			t.Errorf("second sub id = %d, want 1", id2)
		}
	})

	t.Run("same player can hold two subscriptions with distinct channels", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		ch1, id1 := l.AddSubscriber(Player1)
		ch2, id2 := l.AddSubscriber(Player1)
		if id1 == id2 {
			t.Errorf("expected distinct ids, both got %d", id1)
		}
		if ch1 == ch2 {
			t.Error("expected distinct channels for two subscriptions")
		}
	})
}

func TestRemoveSubscriber_cancel(t *testing.T) {
	tests := []struct {
		name           string
		toAdd          []domain.PlayerId
		removeIdx      int
		expectedCancel bool
	}{
		{"human remains", []domain.PlayerId{Player1, Player2}, 0, false},
		{"last human removed no others", []domain.PlayerId{Player1}, 0, true},
		{"last human removed ai present", []domain.PlayerId{Player1, Player1Ai}, 0, true},
		{"ai removed human present", []domain.PlayerId{Player1, Player1Ai}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cancelled := false
			l := &Lobby{subscribers: make(map[int]Subscriber), cancel: func() { cancelled = true }}
			ids := make([]int, len(tt.toAdd))
			for i, p := range tt.toAdd {
				_, ids[i] = l.AddSubscriber(p)
			}
			l.RemoveSubscriber(ids[tt.removeIdx])
			if cancelled != tt.expectedCancel {
				t.Errorf("cancelled = %v, want %v", cancelled, tt.expectedCancel)
			}
		})
	}

	t.Run("nil cancel does not panic", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber), cancel: nil}
		_, id := l.AddSubscriber(Player1)
		l.RemoveSubscriber(id)
	})
}

func TestRemoveSubscriber(t *testing.T) {
	t.Run("removes subscriber and closes channel", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		ch1, id1 := l.AddSubscriber(Player1)
		l.RemoveSubscriber(id1)

		if _, ok := l.subscribers[id1]; ok {
			t.Error("subscriber still present after removal")
		}
		if _, open := <-ch1; open {
			t.Error("channel still open after removal")
		}
	})

	t.Run("removing non-existent id is a no-op", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		l.RemoveSubscriber(999)
	})

	t.Run("only removes target subscriber", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		_, id1 := l.AddSubscriber(Player1)
		_, id2 := l.AddSubscriber(Player2)

		l.RemoveSubscriber(id1)

		if _, ok := l.subscribers[id1]; ok {
			t.Error("removed subscriber still present")
		}
		if _, ok := l.subscribers[id2]; !ok {
			t.Error("other subscriber was incorrectly removed")
		}
	})
}

func Test_getPlayerPosition(t *testing.T) {
	tests := []struct {
		name             string
		positions        []domain.PlayerPosition
		requestedPlayer  domain.PlayerId
		expectedPosition domain.PlayerPositionNumber
		expectedOk       bool
	}{
		{"empty request", []domain.PlayerPosition{}, Player1, 0, false},
		{"working human", []domain.PlayerPosition{{Position: 0, Player: Player1}, {Position: 1, Player: Player2}, {Position: 3, Player: Player1Ai}}, Player1, 0, true},
		{"working ai", []domain.PlayerPosition{{Position: 0, Player: Player1}, {Position: 1, Player: Player2}, {Position: 3, Player: Player1Ai}}, Player1Ai, 3, true},
		{"working miss", []domain.PlayerPosition{{Position: 0, Player: Player1}, {Position: 3, Player: Player1Ai}}, Player2, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			positionToPlayer, err := createPlayerPositionMap(tt.positions)
			if err != nil {
				t.Fatal("error making position to player map")
			}
			lobby := Lobby{
				positionToPlayer: positionToPlayer,
			}
			actualPosition, actualOk := lobby.getPlayerPosition(tt.requestedPlayer)
			if actualOk != tt.expectedOk {
				t.Errorf("expected ok %v but got %v", tt.expectedOk, actualOk)
			}
			if actualPosition != tt.expectedPosition {
				t.Errorf("expected position %d but got %d", tt.expectedPosition, actualPosition)
			}
		})
	}
}

func TestCreateLobby(t *testing.T) {
	gameStub := GameStub{}
	validPositions := []domain.PlayerPosition{
		{Position: 0, Player: Player1},
		{Position: 1, Player: Player2},
	}
	conflictingPositions := []domain.PlayerPosition{
		{Position: 0, Player: Player1},
		{Position: 0, Player: Player2},
	}

	tests := []struct {
		name                      string
		getCurrentPlayerPositions func(ctx context.Context, gameId string) ([]domain.PlayerPosition, error)
		getCurrentEventCount      func(ctx context.Context, gameId string) (int, error)
		expectedErr               error
		expectedLobby             *Lobby
	}{
		{
			"GetCurrentPlayerPositions errors",
			func(ctx context.Context, gameId string) ([]domain.PlayerPosition, error) {
				return nil, domain.ErrDatabase
			},
			func(ctx context.Context, gameId string) (int, error) { return 0, nil },
			domain.ErrDatabase,
			nil,
		},
		{
			"inconsistent positions returns internal error",
			func(ctx context.Context, gameId string) ([]domain.PlayerPosition, error) {
				return conflictingPositions, nil
			},
			func(ctx context.Context, gameId string) (int, error) { return 0, nil },
			domain.ErrInternal,
			nil,
		},
		{
			"GetCurrentEventCount errors",
			func(ctx context.Context, gameId string) ([]domain.PlayerPosition, error) { return validPositions, nil },
			func(ctx context.Context, gameId string) (int, error) { return 0, domain.ErrDatabase },
			domain.ErrDatabase,
			nil,
		},
		{
			"working",
			func(ctx context.Context, gameId string) ([]domain.PlayerPosition, error) { return validPositions, nil },
			func(ctx context.Context, gameId string) (int, error) { return 5, nil },
			nil,
			&Lobby{
				gameId:      "abcde",
				subscribers: map[int]Subscriber{},
				positionToPlayer: map[domain.PlayerPositionNumber]domain.PlayerId{
					0: Player1,
					1: Player2,
				},
				eventCount: 5,
				subCount:   0,
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := GameDbServiceStub{
				getCurrentPlayerPositions: tt.getCurrentPlayerPositions,
				getCurrentEventCount:      tt.getCurrentEventCount,
			}
			lobby, err := CreateLobby(ctx, gameStub, "abcde", db)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("got err %v, want %v", err, tt.expectedErr)
			}
			if tt.expectedLobby != nil {
				opts := []cmp.Option{
					cmp.AllowUnexported(Lobby{}),
					cmpopts.IgnoreFields(Lobby{}, "inputChannel", "game", "dbService", "subMutex"),
				}
				if diff := cmp.Diff(tt.expectedLobby, lobby, opts...); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func assertErrorMessage(t *testing.T, ch <-chan json.RawMessage, wantSeq int, wantMsg string) {
	t.Helper()
	select {
	case msg := <-ch:
		var update GameUpdate
		if err := json.Unmarshal(msg, &update); err != nil {
			t.Errorf("could not unmarshal GameUpdate: %v", err)
			return
		}
		if update.MessageType != domain.ErrorUpdateType {
			t.Errorf("MessageType = %s, want %s", update.MessageType, domain.ErrorUpdateType)
		}
		if update.SequenceNumber != wantSeq {
			t.Errorf("SequenceNumber = %d, want %d", update.SequenceNumber, wantSeq)
		}
		var payload ErrorPayload
		if err := json.Unmarshal(update.Payload, &payload); err != nil {
			t.Errorf("could not unmarshal ErrorPayload: %v", err)
		} else if payload.Message != wantMsg {
			t.Errorf("Message = %q, want %q", payload.Message, wantMsg)
		}
	default:
		t.Error("channel had no message")
	}
}

func Test_sendErrorToUser(t *testing.T) {
	t.Run("sends error only to target player", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		player1Chan, _ := l.AddSubscriber(Player1)
		player2Chan, _ := l.AddSubscriber(Player2)

		l.sendErrorToUser(Player2, 2, "test error")

		assertErrorMessage(t, player2Chan, 2, "test error")
		select {
		case <-player1Chan:
			t.Error("player 1 received message but should not have")
		default:
		}
	})

	t.Run("removes blocked subscriber", func(t *testing.T) {
		blockedChan := make(chan json.RawMessage, 1)
		blockedChan <- []byte("blocker")
		blockedId := 0
		l := &Lobby{
			subscribers: map[int]Subscriber{blockedId: {Player2, blockedChan}},
			subCount:    1,
		}
		openChan, _ := l.AddSubscriber(Player2)

		l.sendErrorToUser(Player2, 1, "test error")

		assertErrorMessage(t, openChan, 1, "test error")
		if _, ok := l.subscribers[blockedId]; ok {
			t.Error("blocked subscriber was not removed")
		}
	})

	t.Run("all open subscriptions for target player receive message", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		ch1, _ := l.AddSubscriber(Player2)
		ch2, _ := l.AddSubscriber(Player2)

		l.sendErrorToUser(Player2, 3, "multi error")

		assertErrorMessage(t, ch1, 3, "multi error")
		assertErrorMessage(t, ch2, 3, "multi error")
	})

	t.Run("player not subscribed is a no-op", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		l.sendErrorToUser(Player1, 0, "test error")
	})
}

func playerIdPayload(t *testing.T, player domain.PlayerId) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(player)
	if err != nil {
		t.Fatalf("failed to marshal player id: %v", err)
	}
	return payload
}

func assertUpdateMessage(t *testing.T, ch <-chan json.RawMessage, wantSeq int, wantPayload json.RawMessage) {
	t.Helper()
	select {
	case msg := <-ch:
		var update GameUpdate
		if err := json.Unmarshal(msg, &update); err != nil {
			t.Errorf("could not unmarshal GameUpdate: %v", err)
			return
		}
		if update.SequenceNumber != wantSeq {
			t.Errorf("SequenceNumber = %d, want %d", update.SequenceNumber, wantSeq)
		}
		if string(update.Payload) != string(wantPayload) {
			t.Errorf("Payload = %s, want %s", update.Payload, wantPayload)
		}
	default:
		t.Error("channel had no message")
	}
}

func Test_broadcast(t *testing.T) {
	payloadFor := func(player domain.PlayerId) json.RawMessage {
		return playerIdPayload(t, player)
	}

	t.Run("all subscribers receive their payload", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		ch1, _ := l.AddSubscriber(Player1)
		ch2, _ := l.AddSubscriber(Player2)

		l.broadcast(1, domain.StateUpdateType, payloadFor)

		assertUpdateMessage(t, ch1, 1, playerIdPayload(t, Player1))
		assertUpdateMessage(t, ch2, 1, playerIdPayload(t, Player2))
	})

	t.Run("removes blocked subscriber", func(t *testing.T) {
		blockedChan := make(chan json.RawMessage, 1)
		blockedChan <- []byte("blocker")
		blockedId := 0
		l := &Lobby{
			subscribers: map[int]Subscriber{blockedId: {Player1, blockedChan}},
			subCount:    1,
		}
		openChan, _ := l.AddSubscriber(Player2)

		l.broadcast(1, domain.StateUpdateType, payloadFor)

		assertUpdateMessage(t, openChan, 1, playerIdPayload(t, Player2))
		if _, ok := l.subscribers[blockedId]; ok {
			t.Error("blocked subscriber was not removed")
		}
	})

	t.Run("invalid payload sends error message to subscriber", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		ch, _ := l.AddSubscriber(Player1)

		l.broadcast(5, domain.StateUpdateType, func(domain.PlayerId) json.RawMessage {
			return json.RawMessage("not valid json")
		})

		assertErrorMessage(t, ch, 5, "missing update")
	})

	t.Run("no subscribers is a no-op", func(t *testing.T) {
		l := &Lobby{subscribers: make(map[int]Subscriber)}
		l.broadcast(0, domain.StateUpdateType, payloadFor)
	})
}

// makeCountingGameStub creates a GameStub backed by an integer counter.
// A move payload of json `true` increments the counter; anything else returns an error.
// GetSaveState/SetGameState serialise and restore the counter for rollback testing.
// GetPlayerState returns the position number as JSON; GetPublicState returns `"public"`.
func makeCountingGameStub(initial int, status string) (*int, GameStub) {
	state := initial
	return &state, GameStub{
		getSaveState: func() json.RawMessage {
			b, _ := json.Marshal(state)
			return b
		},
		setGameState: func(gameState json.RawMessage) {
			_ = json.Unmarshal(gameState, &state)
		},
		makeMove: func(_ domain.PlayerPositionNumber, move json.RawMessage) error {
			var valid bool
			if err := json.Unmarshal(move, &valid); err != nil || !valid {
				return errors.New("invalid move")
			}
			state++
			return nil
		},
		getPublicState: func() json.RawMessage { return json.RawMessage(`"public"`) },
		getPlayerState: func(pos domain.PlayerPositionNumber) json.RawMessage {
			b, _ := json.Marshal(pos)
			return b
		},
		getGameStatus: func() string { return status },
	}
}

// makePassthroughDb builds a GameDbServiceStub whose WithTx executes the callback
// against itself, returning saveStateErr from SaveGameState and saveEventErr from SaveGameEvent.
func makePassthroughDb(saveStateErr, saveEventErr error) GameDbServiceStub {
	var db GameDbServiceStub
	db = GameDbServiceStub{
		withTx: func(ctx context.Context, fn func(GameDbService) error) error {
			return fn(db)
		},
		saveGameState: func(_ context.Context, _ string, _ time.Time, _ json.RawMessage, _ string) error {
			return saveStateErr
		},
		saveGameEvent: func(_ context.Context, _ string, _ domain.PlayerId, _ int, _ string, _ json.RawMessage, _ time.Time) error {
			return saveEventErr
		},
	}
	return db
}

// makeMoveLobby builds a Lobby with Player1 at position 0 and Player2 at position 1.
func makeMoveLobby(game Game, db GameDbService) *Lobby {
	positionToPlayer, _ := createPlayerPositionMap([]domain.PlayerPosition{
		{Position: 0, Player: Player1},
		{Position: 1, Player: Player2},
	})
	return &Lobby{
		subscribers:      make(map[int]Subscriber),
		positionToPlayer: positionToPlayer,
		game:             game,
		dbService:        db,
		eventCount:       3,
	}
}

func Test_handleMoveAction(t *testing.T) {
	ctx := context.Background()
	validMove := json.RawMessage(`true`)
	invalidMove := json.RawMessage(`false`)

	t.Run("player not in position gets error", func(t *testing.T) {
		statePtr, game := makeCountingGameStub(0, domain.OpenGameStatus)
		l := makeMoveLobby(game, GameDbServiceStub{})
		ch, _ := l.AddSubscriber(Player1Ai)

		l.handleMoveAction(ctx, Player1Ai, validMove)

		assertErrorMessage(t, ch, 3, "can not make move without being in the game")
		if *statePtr != 0 {
			t.Errorf("state = %d, want 0 (no move made)", *statePtr)
		}
	})

	t.Run("invalid move gets error", func(t *testing.T) {
		statePtr, game := makeCountingGameStub(0, domain.OpenGameStatus)
		l := makeMoveLobby(game, GameDbServiceStub{})
		ch, _ := l.AddSubscriber(Player1)

		l.handleMoveAction(ctx, Player1, invalidMove)

		assertErrorMessage(t, ch, 3, "tried to make invalid move")
		if *statePtr != 0 {
			t.Errorf("state = %d, want 0 (no move made)", *statePtr)
		}
	})

	t.Run("SaveGameState failure rolls back game state", func(t *testing.T) {
		statePtr, game := makeCountingGameStub(0, domain.OpenGameStatus)
		l := makeMoveLobby(game, makePassthroughDb(domain.ErrDatabase, nil))
		ch, _ := l.AddSubscriber(Player1)

		l.handleMoveAction(ctx, Player1, validMove)

		assertErrorMessage(t, ch, 3, "internal server error")
		if *statePtr != 0 {
			t.Errorf("state = %d, want 0 (rolled back)", *statePtr)
		}
	})

	t.Run("SaveGameEvent failure rolls back game state", func(t *testing.T) {
		statePtr, game := makeCountingGameStub(0, domain.OpenGameStatus)
		l := makeMoveLobby(game, makePassthroughDb(nil, domain.ErrDatabase))
		ch, _ := l.AddSubscriber(Player1)

		l.handleMoveAction(ctx, Player1, validMove)

		assertErrorMessage(t, ch, 3, "internal server error")
		if *statePtr != 0 {
			t.Errorf("state = %d, want 0 (rolled back)", *statePtr)
		}
	})

	t.Run("successful move broadcasts player-specific state", func(t *testing.T) {
		statePtr, game := makeCountingGameStub(0, domain.OpenGameStatus)
		l := makeMoveLobby(game, makePassthroughDb(nil, nil))
		player1Ch, _ := l.AddSubscriber(Player1)
		player2Ch, _ := l.AddSubscriber(Player2)
		nonPlayerCh, _ := l.AddSubscriber(Player1Ai)

		l.handleMoveAction(ctx, Player1, validMove)

		if *statePtr != 1 {
			t.Errorf("state = %d, want 1", *statePtr)
		}
		p1Payload, _ := json.Marshal(domain.PlayerPositionNumber(0))
		assertUpdateMessage(t, player1Ch, 3, p1Payload)
		p2Payload, _ := json.Marshal(domain.PlayerPositionNumber(1))
		assertUpdateMessage(t, player2Ch, 3, p2Payload)
		assertUpdateMessage(t, nonPlayerCh, 3, json.RawMessage(`"public"`))
	})

	t.Run("game status is passed through to SaveGameState", func(t *testing.T) {
		_, game := makeCountingGameStub(0, domain.FinishedGameStatus)
		var capturedStatus string
		var db GameDbServiceStub
		db = GameDbServiceStub{
			withTx: func(ctx context.Context, fn func(GameDbService) error) error {
				return fn(db)
			},
			saveGameState: func(_ context.Context, _ string, _ time.Time, _ json.RawMessage, gameStatus string) error {
				capturedStatus = gameStatus
				return nil
			},
			saveGameEvent: func(_ context.Context, _ string, _ domain.PlayerId, _ int, _ string, _ json.RawMessage, _ time.Time) error {
				return nil
			},
		}
		l := makeMoveLobby(game, db)

		l.handleMoveAction(ctx, Player1, validMove)

		if capturedStatus != domain.FinishedGameStatus {
			t.Errorf("gameStatus = %q, want %q", capturedStatus, domain.FinishedGameStatus)
		}
	})
}
