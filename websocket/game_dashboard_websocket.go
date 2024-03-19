package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type DashboardSignal struct {
	DashboardGame    *DashboardGame     `json:"dashboardGame"`
	DashboardRanking []DashboardRanking `json:"dashboardRanking"`
}

type GameDashboardSocket struct {
	GameDashboardChannel     chan *DashboardSignal
	GetCurrentDashboardState func() (*DashboardSignal, error)
}

type DashboardRanking struct {
	Rank         int    `json:"rank"`
	DisplayName  string `json:"displayName"`
	GamesWon     int    `json:"gamesWon"`
	BestDuration int    `json:"bestDuration"`
}

type DashboardGame struct {
	DocId    string `json:"docId"`
	Duration int    `json:"duration"`

	KiName  string `json:"kiName"`
	KiCoins int    `json:"kiCoins"`

	Player1      string `json:"player1"`
	Player1Coins int    `json:"player1Coins"`

	Player2      string `json:"player2"`
	Player2Coins int    `json:"player2Coins"`

	Player3      string `json:"player3"`
	Player3Coins int    `json:"player3Coins"`

	State string `json:"state"`
}

var (
	dashboardWebsocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (g *GameDashboardSocket) SendToDashboard(dashboardSignal *DashboardSignal) {
	g.GameDashboardChannel <- dashboardSignal
}

func (g *GameDashboardSocket) RemoveGameFromDashboard() {

	state, _ := g.GetCurrentDashboardState()
	state.DashboardGame = nil

	g.GameDashboardChannel <- state
}

func InitGameDashboardSocket(
	dashboardSocketChannel chan *DashboardSignal,
	getCurrentDashboardState func() (*DashboardSignal, error),
) *GameDashboardSocket {

	currentGame, err := getCurrentDashboardState()

	if currentGame != nil && err != nil {
		dashboardSocketChannel <- currentGame
	}

	return &GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: getCurrentDashboardState,
	}
}

func (g *GameDashboardSocket) GameDashboardWebsocketEndpoint() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Printf("start to initialize websocket to dashboard. Try to send current game to dashboard")
		currentGame, err := g.GetCurrentDashboardState()

		if currentGame != nil && err == nil {
			g.SendToDashboard(currentGame)
		}

		ws, err := dashboardWebsocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade angular user ui websocket error %s", err)
		}

		done := make(chan struct{})

		go gameDashboardWriter(ws, done, g.GameDashboardChannel)
		go gameDashboardReader(ws, done)
	}
}

func gameDashboardWriter(conn *websocket.Conn, done chan struct{}, announcedGameChannel chan *DashboardSignal) {
	defer conn.Close()
	for {
		select {
		case <-done:
			return
		case message := <-announcedGameChannel:

			marshalledGame, err := json.Marshal(&message)

			if err != nil {
				log.Printf("marshal game to json failed: %s", err)
				return
			}

			err = conn.WriteMessage(1, marshalledGame)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func gameDashboardReader(conn *websocket.Conn, done chan struct{}) {
	defer conn.Close()
	defer close(done)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("reader angular user ui websocket error: %s", err)
			return
		}
	}
}
