package shortener

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/w32blaster/shortana/db"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
)

var (
	tmplIndex = template.Must(template.ParseFiles("templates/index.html"))
)

type (
	AllLinksData struct {
		Links    []db.ShortURL
		Error    error
		WrongUrl string
		Hostname string
	}
)

func makeRequestProcessor(db *db.Database) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("fdfd"))
	}
}

// printIndex prints page with available public (!) links in case if short URL was wrong
func printIndex(db *db.Database, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	links, err := db.GetAll()
	data := AllLinksData{
		Links:    links,
		Error:    err,
		Hostname: "http://localhost:3000",
	}

	if err = tmplIndex.Execute(w, data); err != nil {
		log.Println("Error while rendering page: " + err.Error())
	}
}

// StartServer starts the server that handles all the requests
func StartServer(db *db.Database) {

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		printIndex(db, w)
	})
	r.Get("/*", makeRequestProcessor(db))

	http.ListenAndServe(":3000", r)
}
