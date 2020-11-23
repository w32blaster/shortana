package shortener

import (
	"github.com/go-chi/httprate"
	"net/http"
	"strings"
	"time"

	"github.com/w32blaster/shortana/db"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func makeRequestProcessor(db *db.Database) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("fdfd"))
	}
}

// printIndex prints page with available public (!) links in case if short URL was wrong
func printIndex(db *db.Database, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	var sb strings.Builder
	sb.WriteString("<html><body><h1>Shortana URL shortener</h1> <p> Available URLs: <li>")

	allURLs := db.GetAll()
	for k, v := range allURLs {
		sb.WriteString("<ul><a href='")
		sb.WriteString(v)
		sb.WriteString("'>")
		sb.WriteString(k)
		sb.WriteString("</a></ul>")
	}
	sb.WriteString("</li></body></html>")
	w.Write([]byte(sb.String()))
}

// StartServer starts the server that handles all the requests
func StartServer(db *db.Database) {

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Enable httprate request limiter of 100 requests per minute.
	//
	// In the code example below, rate-limiting is bound to the request IP address
	// via the LimitByIP middleware handler.
	//
	// To have a single rate-limiter for all requests, use httprate.LimitAll(..).
	//
	// Please see _example/main.go for other more, or read the library code.
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		printIndex(db, w)
	})
	r.Get("/*", makeRequestProcessor(db))

	http.ListenAndServe(":3000", r)
}
