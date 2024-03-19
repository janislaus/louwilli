package louie_kafka

type EventType string

const (
	PlayersCanBeReceived EventType = "PLAYERS_CAN_BE_RECEIVED"
	PlayersReady         EventType = "PLAYERS_READY"
	PlayersConfirm       EventType = "PLAYERS_CONFIRM"
	GameDone             EventType = "GAME_DONE"
	CoinDrop             EventType = "COIN_DROP"

	PleaseChangeSide     EventType = "PLZ_CHANGE_SIDE"
	ConfirmedChangedSide EventType = "CONFIRMED_CHANGE_SIDE"
	ResetGame            EventType = "RESET_GAME"
)

func (c EventType) String() string {
	return string(c)
}

func getTechnicalEventTypes() []EventType {
	return []EventType{ConfirmedChangedSide, PleaseChangeSide}
}

func getGameEventTypes() []EventType {
	return []EventType{PlayersCanBeReceived, PlayersReady, PlayersConfirm, GameDone, CoinDrop}
}

type GameDoneEvent struct {
	Event         EventType     `json:"event"`
	Sender        string        `json:"sender"`
	Duration      float64       `json:"duration"`
	WinningPlayer winningPlayer `json:"winning_player"`
}

type winningPlayer struct {
	Name string `json:"name"`
}

type PlayersReadyEvent struct {
	Event     EventType           `json:"event"`
	Players   []PlayerDisplayName `json:"players"`
	Timestamp string              `json:"timestamp"`
}

type PlayerDisplayName struct {
	DisplayName string `json:"display_name"`
}

type CoinDropEvent struct {
	Event  EventType `json:"event"`
	Sender string    `json:"sender"`
	Name   string    `json:"name"`
	Coins  int       `json:"coins"`
}

type DefaultEvent struct {
	Event EventType `json:"event"`
}

type resetGame struct {
	Event  EventType `json:"event"`
	Sender string    `json:"sender"`
}

type playersConfirm struct {
	Event  EventType `json:"event"`
	Sender string    `json:"sender"`
}

type PlayersCanBeReceivedEvent struct {
	Event  EventType `json:"event"`
	Sender string    `json:"sender"`
}
