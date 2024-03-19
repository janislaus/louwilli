package admin

import (
	"louie-web-administrator/service"
	"net/http"
)

func ConfirmSideChange(technicalEventHandler *service.TechnicalEventHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		technicalEventHandler.SendConfirmedChangeSideEvent()

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(200)

		_, _ = w.Write([]byte("<div hx-swap-oob=\"replace:#confirm-change-side\"></div>"))
	}
}
