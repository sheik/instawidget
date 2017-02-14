package main

import (
	"net/http"
	"log"
	"html/template"
	"fmt"

	"github.com/gorilla/mux"
)

// serve up an error file
// this shit needs some review
func serveError(w http.ResponseWriter, r *http.Request, status int) {
	log.Printf("HTTP %d: %s", status, r.URL.Path)
	t, err := template.ParseFiles(fmt.Sprintf("%s%d.html", TEMPLATE_DIR, status))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// staticHandler handles all unconfigured paths, serving the
// files out of the directory defined in STATIC_DIR
func staticHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := STATIC_DIR + vars["filename"]

	if len(path) > 0 && path[len(path)-1] == '/' {
		path += "index.html"
	}

	// serve file if it exists
	if pathExists(path) {
		http.ServeFile(w, r, path)
		return
	}

	serveError(w, r, http.StatusNotFound)
}