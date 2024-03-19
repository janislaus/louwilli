package dashboard

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
	"louie-web-administrator/service"
	"net/http"
	"strings"
)

type dashboardUserRequest struct {
	User dashboardUser `json:"user"`
}
type dashboardUser struct {
	AcceptNewsletter   bool   `json:"acceptNewsletter"`
	AcceptNotification bool   `json:"acceptNotification"`
	DisplayName        string `json:"displayName" validate:"required"`
	Email              string `json:"email" validate:"required,email"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
}

func Post(userService *service.UserSer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var request dashboardUserRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Printf("json decoder init failed: %s for %s\n", err, r.Body)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		validate := validator.New(validator.WithRequiredStructEnabled())
		err := validate.Struct(request)
		if err != nil {
			log.Printf("user request contains failures %s", err)
			http.Error(w, fmt.Sprintf("failed to parse user input %s", err), http.StatusBadRequest)
			return
		}

		_, err = userService.Create(&service.DashboardUser{
			AcceptNewsletter:   request.User.AcceptNewsletter,
			AcceptNotification: request.User.AcceptNotification,
			DisplayName:        strings.ToLower(strings.TrimSpace(request.User.DisplayName)),
			Email:              strings.ToLower(strings.TrimSpace(request.User.Email)),
			FirstName:          strings.ToLower(strings.TrimSpace(request.User.FirstName)),
			LastName:           strings.ToLower(strings.TrimSpace(request.User.LastName)),
		})

		if err != nil {

			if strings.Contains(err.Error(), "E11000") {
				log.Printf("duplicate %s", err)
				http.Error(w, fmt.Sprintf("duplicate user: %s", "E11000"), http.StatusBadRequest)
				return
			} else {
				log.Printf("creating user failed %s", err)
				http.Error(w, fmt.Sprintf("creating user failed %s", err), http.StatusInternalServerError)
				return
			}
		}

		log.Printf("created user successfully %s", request.User.Email)

		w.WriteHeader(http.StatusCreated)
	}
}
