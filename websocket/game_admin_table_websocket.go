package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	_ "github.com/gorilla/websocket"
	"log"
	"net/http"
)

type AdminUiEventType string

const (
	Announced                 AdminUiEventType = "announced"
	Ready                     AdminUiEventType = "ready"
	Active                    AdminUiEventType = "active"
	Finished                  AdminUiEventType = "finished"
	PlzChangeSide             AdminUiEventType = "plz_change_side"
	ActivateGameStartButton                    = "activate_game_start"
	DeactivateGameStartButton                  = "deactivate_game_start"
)

func (c AdminUiEventType) String() string {
	return string(c)
}

func ValueOf(value string) (AdminUiEventType, error) {
	if value == Announced.String() {
		return Announced, nil
	} else if value == Ready.String() {
		return Ready, nil
	} else if value == Active.String() {
		return Active, nil
	} else if value == Finished.String() {
		return Finished, nil
	} else if value == PlzChangeSide.String() {
		return PlzChangeSide, nil
	} else {
		log.Printf("mapping of admin ui event type failed")
		return PlzChangeSide, nil
	}
}

type AdminUiEvent struct {
	EventType    AdminUiEventType
	KiCoins      int
	Player1Coins int
	Player2Coins int
	Player3Coins int
}

type AdminUiWebsocket struct {
	adminUiChannel chan AdminUiEvent
}

var AdminUiWebsocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (a *AdminUiWebsocket) SendToAdminUi(adminUiEvent *AdminUiEvent) {
	a.adminUiChannel <- *adminUiEvent
}

func InitAdminUiWebsocket(adminUiChannel chan AdminUiEvent) *AdminUiWebsocket {
	return &AdminUiWebsocket{adminUiChannel: adminUiChannel}
}

func (a *AdminUiWebsocket) AdminUiWebsocketEndpoint() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		ws, err := AdminUiWebsocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade error %s", err)
		}

		done := make(chan struct{})
		go adminUiWriter(ws, done, a.adminUiChannel)
		go adminUiReader(ws, done)
	}
}

func adminUiWriter(conn *websocket.Conn, done chan struct{}, adminUiChannel chan AdminUiEvent) {
	defer conn.Close()
	for {
		select {
		case <-done:
			return
		case message := <-adminUiChannel:

			gameStateHtmlSnippet := createGameStateHtmlSnippet(message)

			err := conn.WriteMessage(1, []byte(gameStateHtmlSnippet))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func createGameStateHtmlSnippet(adminUiSignal AdminUiEvent) string {
	var renderedMessage string

	switch adminUiSignal.EventType {
	case PlzChangeSide:
		renderedMessage = fmt.Sprintf("<div hx-swap-oob=\"replace:#confirm-change-side\">" +
			"<button class=\"btn btn-secondary\" hx-post=\"/confirm\">Confirm side change</button>" +
			"</div>")
	case Announced:
		renderedMessage = fmt.Sprintf("<div hx-swap-oob=\"replace:#game-state\"><p class=\"state-announced\">%s</p></div>", adminUiSignal.EventType)
	case Ready:
		renderedMessage = fmt.Sprintf("<div hx-swap-oob=\"replace:#game-state\"><p class=\"state-ready\">!!!! %s !!!!</p></div>", adminUiSignal.EventType)
	case Active:
		renderedMessage = fmt.Sprintf(""+
			"<div hx-swap-oob=\"replace:#game-state\"><p class=\"state-active\">!!!! %s !!!!</p></div>"+
			"<div hx-swap-oob=\"replace:#ki-coins\"><p>%d</p></div>"+
			"<div hx-swap-oob=\"replace:#player1-coins\"><p>%d</p></div>"+
			"<div hx-swap-oob=\"replace:#player2-coins\"><p>%d</p></div>"+
			"<div hx-swap-oob=\"replace:#player3-coins\"><p>%d</p></div>"+
			"", adminUiSignal.EventType, adminUiSignal.KiCoins, adminUiSignal.Player1Coins, adminUiSignal.Player2Coins, adminUiSignal.Player3Coins)
	case Finished:
		renderedMessage = fmt.Sprintf("<div hx-swap-oob=\"replace:#game-state\"><p class=\"state-finished\">!!!! %s !!!!</p></div>", adminUiSignal.EventType)
	}

	return renderedMessage
}

func adminUiReader(conn *websocket.Conn, done chan struct{}) {
	defer conn.Close()
	defer close(done)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("reader websocket error (maybe side is refreshed): %s", err)
			return
		}
	}
}
