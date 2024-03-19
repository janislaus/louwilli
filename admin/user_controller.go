package admin

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"louie-web-administrator/service"
	"net/http"
	"strconv"
)

func Page(userService *service.UserSer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		page, err := strconv.ParseInt(params["page"], 10, 16)
		nameFilter := r.URL.Query().Get("name-filter")

		if err != nil {
			log.Printf("can not convert page parameter: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		usersTemplate, err := generateUserTemplateContent(userService, page, nameFilter)

		if err != nil {
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(200)

		_, err = w.Write(usersTemplate.Bytes())

		if err != nil {
			log.Printf("writing registered users template to output writer failed %s\n", err)
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}
	}
}
func Position(userService *service.UserSer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if _, ok := parseForm(w, r); ok {
			return
		}

		userIdsToPositions, err := userService.MapUserIdsToPositions(r.Form)
		pageNumber := readPageNumberFromRequest(r.Form)
		nameFilter := readFilterValueFromRequest(r.Form)

		if err != nil {
			log.Printf("parsing of user ids to positions failed: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, userIdToPosition := range userIdsToPositions {
			_, err := userService.UpdatePosition(userIdToPosition.Key, userIdToPosition.Value)

			if err != nil {
				log.Printf("updating user %s position failed: %s\n", userIdToPosition.Value, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		usersTemplate, err := generateUserTemplateContent(userService, pageNumber, nameFilter)

		if err != nil {
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(200)

		_, err = w.Write(usersTemplate.Bytes())

		if err != nil {
			log.Printf("writing registered users template to output writer failed %s\n", err)
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}
	}
}

func Shuffle(userService *service.UserSer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if _, ok := parseForm(w, r); ok {
			return
		}

		pageNumber := readPageNumberFromRequest(r.Form)
		nameFilter := readFilterValueFromRequest(r.Form)

		userService.ShufflePositionsForActiveUsers()

		usersTemplate, err := generateUserTemplateContent(userService, pageNumber, nameFilter)

		if err != nil {
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(200)

		_, err = w.Write(usersTemplate.Bytes())

		if err != nil {
			log.Printf("writing registered users template to output writer failed %s\n", err)
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}
	}
}

func State(userService *service.UserSer, adminEventService *service.AdminEventService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if _, ok := parseForm(w, r); ok {
			return
		}

		userIdsToStates, err := userService.MapUserIdsToStates(r.Form)
		pageNumber := readPageNumberFromRequest(r.Form)
		nameFilter := readFilterValueFromRequest(r.Form)

		if err != nil {
			log.Printf("parsing of user ids to states failed: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, userIdToState := range userIdsToStates {
			_, err := userService.UpdateState(userIdToState.Key, userIdToState.Value)

			if err != nil {
				log.Printf("updating user %s state failed: %s\n", userIdToState.Value, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		adminEventService.CheckActiveUsersAndEnableOrDisableGameButton()

		usersTemplate, err := generateUserTemplateContent(userService, pageNumber, nameFilter)

		if err != nil {
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(200)

		_, err = w.Write(usersTemplate.Bytes())

		if err != nil {
			log.Printf("writing registered users template to output writer failed %s\n", err)
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}
	}
}

func Wait(userService *service.UserSer, adminEventService *service.AdminEventService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if _, ok := parseForm(w, r); ok {
			return
		}

		pageNumber := readPageNumberFromRequest(r.Form)
		nameFilter := readFilterValueFromRequest(r.Form)

		err := userService.SetAllToWaiting()

		if err == nil {
			adminEventService.CheckActiveUsersAndEnableOrDisableGameButton()
		}

		usersTemplate, err := generateUserTemplateContent(userService, pageNumber, nameFilter)

		if err != nil {
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(200)

		_, err = w.Write(usersTemplate.Bytes())

		if err != nil {
			log.Printf("writing registered users template to output writer failed %s\n", err)
			http.Error(w, fmt.Sprintf("something goes wrong during rendering the registered users template %s", err), http.StatusInternalServerError)
			return
		}
	}
}

func generateUserTemplateContent(userService *service.UserSer, page int64, nameFilter string) (*bytes.Buffer, error) {

	pagedUsers := userService.GetUsers(page, nameFilter)
	activeUsersCount := userService.CountActiveUsersWithoutKiUser()
	usersTemplate, err := renderRegisteredUsersTemplate(pagedUsers, calculatePages(userService.CountAllWithoutKiUser(nameFilter), page), activeUsersCount, nameFilter)

	return usersTemplate, err
}
func parseForm(w http.ResponseWriter, r *http.Request) (error, bool) {
	err := r.ParseForm()

	if err != nil {
		log.Printf("parse form failed: %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, true
	}

	return err, false
}

func readPageNumberFromRequest(formParameters map[string][]string) int64 {

	defaultPageNumber := int64(1)

	if formParameters["page-number"] != nil {
		number, _ := strconv.Atoi(formParameters["page-number"][0])
		defaultPageNumber = int64(number)
	}

	return defaultPageNumber
}

func renderRegisteredUsersTemplate(users []service.UserEntry, paging paging, activeUsersCount int, nameFilterValue string) (*bytes.Buffer, error) {

	var output bytes.Buffer

	tmpl, err := mainTemplate()

	if err != nil {
		log.Printf("can not render registered users template %s\n", err)
		return nil, err
	}

	err = tmpl.ExecuteTemplate(&output, "users-table", templateContent{
		UserEntries:      users,
		Paging:           paging,
		ActiveUsersCount: activeUsersCount,
		NameFilter:       nameFilterValue,
	})

	if err != nil {
		log.Printf("generate registered users template failed %s\n", err)
		return nil, err
	}

	return &output, nil
}
