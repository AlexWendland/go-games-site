package domain

const (
	// Cookies
	SessionCookieName           = "session"
	UserIDCookieName            = "user_id"
	SessionExpriationCookieName = "session_expires_at"

	// Game event keys in DB
	MoveGameEvent  = "move"
	LobbyGameEvent = "lobby"
	AiGameEvent    = "ai"

	// Update types in json
	StateUpdateType = "state"
	LobbyUpdateType = "lobby"
	ErrorUpdateType = "error"

	// Game status
	OpenGameStatus     = "open"
	FinishedGameStatus = "finished"
)
