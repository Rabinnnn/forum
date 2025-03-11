package handlers

import (
	"html/template"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/github.html")
	tmpl.Execute(w, nil)
}
