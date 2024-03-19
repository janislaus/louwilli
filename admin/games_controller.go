package admin

import (
	"bytes"
	"fmt"
	"log"
	"louie-web-administrator/service"
	"louie-web-administrator/websocket"
	"net/http"
)

func RemoveGame(userService *service.UserSer, gameService *service.GameSer, gameDashboardSocket *websocket.GameDashboardSocket) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		_, err := gameService.RemoveGame(r.Header.Get("Hx-Trigger"))

		if err != nil {
			http.Error(w, fmt.Sprintf("remove game failed %s", err), http.StatusInternalServerError)
			return
		}

		gameDashboardSocket.RemoveGameFromDashboard()

		gamesTemplate, err := renderGameTemplate(userService, gameService)

		if err != nil {
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the games template %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(200)

		_, err = w.Write(gamesTemplate.Bytes())

		if err != nil {
			log.Printf("writing games template to output writer failed %s\n", err)
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the games template %s", err), http.StatusInternalServerError)
			return
		}
	}
}

func AnnounceGame(userService *service.UserSer, gameService *service.GameSer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		activeUsers, err := userService.GetAllActive()

		if err != nil {
			http.Error(w, fmt.Sprintf("can not find active registered users %s", err), http.StatusInternalServerError)
			return
		}

		_, err = gameService.CreateGame(activeUsers)

		if err != nil {
			http.Error(w, fmt.Sprintf("announcing game failed %s", err), http.StatusInternalServerError)
			return
		}

		gamesTemplate, err := renderGameTemplate(userService, gameService)

		if err != nil {
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the games template %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(200)

		_, err = w.Write(gamesTemplate.Bytes())

		if err != nil {
			log.Printf("writing games template to output writer failed %s\n", err)
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the games template %s", err), http.StatusInternalServerError)
			return
		}
	}
}
func renderGameTemplate(userService *service.UserSer, gameService *service.GameSer) (*bytes.Buffer, error) {

	var output bytes.Buffer

	tmpl, err := mainTemplate()

	if err != nil {
		log.Printf("can not render games template %s\n", err)
		return nil, err
	}

	game, err := gameService.GetCurrentGame()
	if err != nil {
		log.Printf("get games failed: %s\n", err)
		return nil, err
	}

	if game == nil {
		err = tmpl.ExecuteTemplate(&output, "games-table-content", templateContent{GameEntries: []service.GameEntry{}, ActiveUsersCount: userService.CountActiveUsersWithoutKiUser()})
	} else {
		err = tmpl.ExecuteTemplate(&output, "games-table-content", templateContent{GameEntries: []service.GameEntry{*game}, ActiveUsersCount: userService.CountActiveUsersWithoutKiUser()})
	}

	if err != nil {
		log.Printf("generate games template failed %s\n", err)
		return nil, err
	}

	return &output, nil
}
