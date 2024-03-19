package admin

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"louie-web-administrator/service"
	"math"
	"net/http"
)

//go:embed *.gohtml
var templates embed.FS

const (
	MainTemplate   = "index.gohtml"
	UserTemplate   = "user.gohtml"
	GameTemplate   = "game.gohtml"
	PagingTemplate = "paging.gohtml"
)

type templateContent struct {
	UserEntries      []service.UserEntry
	GameEntries      []service.GameEntry
	Paging           paging
	ActiveUsersCount int
	NameFilter       string
}

type paging struct {
	Pages []page
}

type page struct {
	Number int64
	Active bool
}

func Main(userService *service.UserSer, gameService *service.GameSer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		writeMainTemplate(w, userService, gameService)
	}
}

func writeMainTemplate(w http.ResponseWriter, userService *service.UserSer, gameService *service.GameSer) {
	mainTemplateContent, err := renderMainTemplate(userService, gameService)

	if err != nil {
		http.Error(w, fmt.Sprintf("something goes wrong during rendering the admin main template %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/html")
	w.WriteHeader(200)

	_, err = w.Write(mainTemplateContent.Bytes())

	if err != nil {
		log.Printf("writing main template to output writer failed %s\n", err)
		http.Error(w, fmt.Sprintf("something goes wrong during rendering the admin main template %s", err), http.StatusInternalServerError)
		return
	}
}

func renderMainTemplate(userService *service.UserSer, gameService *service.GameSer) (*bytes.Buffer, error) {

	var output bytes.Buffer

	tmpl, err := mainTemplate()

	if err != nil {
		log.Printf("can not render main template %s\n", err)
		return nil, err
	}

	pagedUsers := userService.GetUsers(1, "")
	activeUsersCount := userService.CountActiveUsersWithoutKiUser()
	game, err := gameService.GetCurrentGame()

	if err != nil {
		return nil, err
	}

	templateContent := templateContent{
		UserEntries:      pagedUsers,
		GameEntries:      []service.GameEntry{},
		Paging:           calculatePages(userService.CountAllWithoutKiUser(""), 1),
		ActiveUsersCount: activeUsersCount,
		NameFilter:       "",
	}

	if game == nil {
		templateContent.GameEntries = []service.GameEntry{}
	} else {
		templateContent.GameEntries = []service.GameEntry{*game}
	}

	err = tmpl.Execute(&output, templateContent)

	if err != nil {
		log.Printf("generate main template failed %s\n", err)
		return nil, err
	}

	return &output, nil
}

func mainTemplate() (*template.Template, error) {
	tmpl, err := template.ParseFS(templates, MainTemplate, UserTemplate, GameTemplate, PagingTemplate)

	return tmpl, err
}

func calculatePages(sizeOfAllUsers, activePage int64) paging {

	var sizeOfPages int64

	integerPart, _ := math.Modf(float64(sizeOfAllUsers / 10))
	existsRestOfDivision := (sizeOfAllUsers % 10) > 0

	if existsRestOfDivision {
		sizeOfPages = int64(integerPart) + 1
	} else {
		sizeOfPages = int64(integerPart)
	}

	pages := make([]page, 0, sizeOfPages)

	for i := int64(1); i <= sizeOfPages; i++ {
		if i == activePage {
			pages = append(pages, page{
				Number: i,
				Active: true,
			})
		} else {
			pages = append(pages, page{
				Number: i,
				Active: false,
			})
		}
	}

	return paging{Pages: pages}
}
