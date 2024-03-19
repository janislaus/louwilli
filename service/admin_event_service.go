package service

import "louie-web-administrator/websocket"

type AdminEventService struct {
	gameService    GameService
	userService    UserService
	adminWebsocket *websocket.AdminUiWebsocket
}

func InitAdminEventService(userService UserService, gameService GameService, adminWebsocket *websocket.AdminUiWebsocket) *AdminEventService {
	return &AdminEventService{
		userService:    userService,
		gameService:    gameService,
		adminWebsocket: adminWebsocket,
	}
}

func (a *AdminEventService) CheckActiveUsersAndEnableOrDisableGameButton() {

	activeUsersCount := a.userService.CountActiveUsersWithoutKiUser()
	currentGame, _ := a.gameService.GetCurrentGame()

	if activeUsersCount >= 3 && currentGame == nil {
		a.adminWebsocket.SendToAdminUi(&websocket.AdminUiEvent{
			EventType: websocket.ActivateGameStartButton,
		})
	} else {
		a.adminWebsocket.SendToAdminUi(&websocket.AdminUiEvent{
			EventType: websocket.DeactivateGameStartButton,
		})
	}
}
