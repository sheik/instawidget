package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"code.google.com/p/gosqlite/sqlite"
	"github.com/gorilla/mux"
)

type Page struct {
	Title    string
	Body     template.HTML
	Response OAuthResponse
}

// User information returned
// by instagram REST API
type User struct {
	Username        string
	Bio             string
	Website         string
	Profile_Picture string
	Full_Name       string
	Id              string
}

// OAuthResponse provides the evelope for use in
// obtaining an access token for a user
type OAuthResponse struct {
	Access_Token string
	User         User
}

// Instagram Envelope
// http://instagram.com/developer/endpoints/#
//
// The full envelope is not implemented here,
// only the parts that are used have been code
type Envelope struct {
	Meta struct {
		Code int
	}
	Pagination struct {
		Next_url    string
		Next_max_id string
	}
	Data []struct {
		Link   string
		User   User
		Images map[string]struct {
			Url    string
			Width  int
			Height int
		}
	}
}

// pathExists check if the file or directory specified by
// the argument path exists and returns false if it does
// not exist or true otherwise
func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.URL.Query().Get("user_id")

	// connect to the database
	db, err := sqlite.Open(DBFILE)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("select access_token from users where id = ?")
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Exec(user_id)
	if err != nil {
		log.Fatal(err)
	}

	var access_token string

	if !stmt.Next() {
		return
	}

	stmt.Scan(&access_token)

	err = stmt.Finalize()
	if err != nil {
		log.Fatal(err)
	}

	endpoint := "https://api.instagram.com/v1/users/" + user_id + "/media/recent/?count=6&access_token=" + access_token

	resp, err := http.Get(endpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var e Envelope
	err = json.Unmarshal(body, &e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := "<div style='width:250px'>"

	for _, t := range e.Data {
		p += fmt.Sprintf("<a href='%s' target='_blank'><img src='%s' style='width:120px;'/></a> ", t.Link, t.Images["thumbnail"].Url)
	}

	p += "</div>"

	t := *buildTemplate()

	if err := t["index"].ExecuteTemplate(w, "base", Page{Title: "nerd", Body: template.HTML(p)}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// staticHandler handles all unconfigured paths, serving the
// files out of the directory defined in STATIC_DIR
func OAthHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	resp, err := http.PostForm("https://api.instagram.com/oauth/access_token", url.Values{
		"client_id":     {INSTAGRAM_CLIENT_ID},
		"client_secret": {INSTAGRAM_CLIENT_SECRET},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {REDIRECT_URI},
		"code":          {code},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var oar OAuthResponse
	err = json.Unmarshal(body, &oar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// connect to the database
	db, err := sqlite.Open(DBFILE)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Exec("insert or ignore into users (id, access_token) values (?, ?)", oar.User.Id, oar.Access_Token)
	if err != nil {
		panic(err)
	}

	t := *buildTemplate()

	err = t["welcome"].ExecuteTemplate(w, "base", Page{Title: "nerd", Response: oar})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func buildTemplate() *map[string]*template.Template {
	t := make(map[string]*template.Template)
	t["index"] = template.Must(template.ParseFiles(TEMPLATE_DIR+"base.html", TEMPLATE_DIR+"basic.html"))
	t["welcome"] = template.Must(template.ParseFiles(TEMPLATE_DIR+"base.html", TEMPLATE_DIR+"welcome.html"))
	return &t
}

func registerIndex(w http.ResponseWriter, r *http.Request) {
	message := "You must authenticate <a href='" + INSTAGRAM_OAUTH + "'>here</a>."
	t := *buildTemplate()

	err := t["index"].ExecuteTemplate(w, "base", Page{Title: "nerd", Body: template.HTML(message)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// initializeDatabase inits the database structure for the
// application. It is evoked automatically if the database
// file does not exist, or forcibly by the user
func initializeDatabase(conn *sqlite.Conn) {
	structure := `
	create table users (
		id int not null,
		access_token text,
		primary key(id),
		unique(access_token)
	)
	`
	if err := conn.Exec(structure); err != nil {
		log.Fatal(err)
	}
}

func main() {

	address := flag.String("address", "127.0.0.1", "Address to listen on")
	port := flag.String("port", "9999", "Port to listen on")
	initdb := flag.Bool("newdb", false, "Re-initialize the database")

	flag.Parse()

	// initialize database if the db file
	// does not exist
	if !pathExists(DBFILE) {
		*initdb = true
	}

	// connect to the database
	db, err := sqlite.Open(DBFILE)
	if err != nil {
		log.Fatal(err)
	}

	// initialize the database if required
	if *initdb {
		initializeDatabase(db)
	}

	db.Close()

	// set up routes
	r := mux.NewRouter()
	r.HandleFunc("/iws/{rest:.*}", userHandler)
	r.HandleFunc("/iw/{filename:.*}", OAthHandler)
	r.HandleFunc("/{rest:.*}", registerIndex)
	http.Handle("/", r)

	fmt.Printf("Launching %s on https://%s:%s\n", PROGRAM_VERSION, *address, *port)

	if err := http.ListenAndServeTLS(*address+":"+*port, "/path/to/cert.pem", "/path/to/server.key", nil); err != nil {
		log.Fatal(err)
	}
}
