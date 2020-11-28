package shortener

import (
	"github.com/w32blaster/shortana/stats"
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

func makeRequestProcessor(db *db.Database, stats *stats.Statistics, hostname string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		shortUrl := chi.URLParam(req, "shortUrl")
		if len(shortUrl) == 0 {
			w.Write([]byte("Short URL is not provided"))
			return
		}

		url, err := db.GetUrl(shortUrl)
		if err != nil {
			printIndex(db, w, hostname, shortUrl)
			return
		}

		go stats.ProcessRequest(req, shortUrl)

		w.Header().Add("Location", url.TargetUrl)
		w.WriteHeader(http.StatusMovedPermanently)
	}
}

// printIndex prints page with available public (!) links in case if short URL was wrong
func printIndex(db *db.Database, w http.ResponseWriter, hostname, wrongUrl string) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	links, err := db.GetAll()
	data := AllLinksData{
		Links:    links,
		Error:    err,
		Hostname: hostname,
		WrongUrl: wrongUrl,
	}

	if err = tmplIndex.Execute(w, data); err != nil {
		log.Println("Error while rendering page: " + err.Error())
	}
}

// StartServer starts the server that handles all the requests
func StartServer(db *db.Database, stats *stats.Statistics, host string) {

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.NoCache)
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		printIndex(db, w, host, "")
	})
	r.Get("/{shortUrl}", makeRequestProcessor(db, stats, host))

	http.ListenAndServe(":3000", r)
}
