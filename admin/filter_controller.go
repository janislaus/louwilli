package admin

import (
	"fmt"
	"log"
	"louie-web-administrator/service"
	"net/http"
)

func Filter(userService *service.UserSer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if _, ok := parseForm(w, r); ok {
			return
		}

		nameFilterValue := readFilterValueFromRequest(r.Form)

		pagedUsers := userService.GetUsers(1, nameFilterValue)
		activeUsersCount := userService.CountActiveUsersWithoutKiUser()

		usersTemplate, err := renderRegisteredUsersTemplate(pagedUsers, calculatePages(userService.CountAllWithoutKiUser(nameFilterValue), 1), activeUsersCount, nameFilterValue)

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

func readFilterValueFromRequest(formParameters map[string][]string) string {

	defaultFilter := ""

	if formParameters["name-filter"] != nil {
		defaultFilter = formParameters["name-filter"][0]
	}

	return defaultFilter
}
