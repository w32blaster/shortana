package main

import (
	"fmt"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/w32blaster/shortana/bot"
	"github.com/w32blaster/shortana/db"
	"github.com/w32blaster/shortana/shortener"
	"github.com/w32blaster/shortana/stats"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/msgpack"
	flags "github.com/jessevdk/go-flags"
	"go.etcd.io/bbolt"
)

// Opts command line arguments
type Opts struct {
	Port              int    `short:"p" long:"port" description:"The port for the bot. The default is 8444" default:"8444"`
	Host              string `short:"h" long:"host" description:"The full hostname for the Shortana. Default is localhost" default:"http://localhost:3000"`
	BotToken          string `short:"b" long:"bot-token" description:"The Bot-Token. As long as it is the sensitive data, we can't keep it in Github" required:"true"`
	AcceptFromUser    int    `short:"u" long:"accept-from-user" description:"Telegram User Id bot can only speak to. By default it can talk to everyone" required:"false"`
	IsDebug           bool   `short:"d" long:"debug" description:"Is it debug? Default is true. Disable it for production."`
	GeoIpDatabasePath string `short:"g" long:"geoip-path" description:"The path to GeoLite database mmbd" default:"GeoLite2-City.mmdb"`
}

func main() {
	fmt.Println("Start the Shortana")

	// parse flags
	var opts = Opts{}
	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}

	// open the GeoIP database
	geoIPdb, err := geoip2.Open(opts.GeoIpDatabasePath)
	if err != nil {
		panic(err)
	}
	defer geoIPdb.Close()

	// Open Storm DB
	boltdb, err := storm.Open("shortana.db", storm.Codec(msgpack.Codec), storm.BoltOptions(0600, &bbolt.Options{Timeout: 5 * time.Second}))
	if err != nil {
		panic(err)
	}
	defer boltdb.Close()

	database := db.Init(boltdb)

	// for development only
	saveDummyLink(database, "STORM", "https://github.com/asdine/storm#options", "Storm project at GitHub", true)
	saveDummyLink(database, "yeti", "https://www.ebay.co.uk/itm/Blue-Yeti-Professional-Multi-Pattern-USB-Mic-for-Record-and-Stream-Sunset-Sky/274582563719?_trkparms=aid%3D777001%26algo%3DDISCO.FEED%26ao%3D1%26asc%3D20200211172457%26meid%3Dd44c41c9188544f0b2a06ab469877571%26pid%3D101213%26rk%3D1%26rkt%3D1%26mehot%3Dnone%26itm%3D274582563719%26pmt%3D0%26noa%3D1%26pg%3D2380057%26algv%3DRecommendingSearch%26brand%3DBlue+Microphones&_trksid=p2380057.c101213.m46344&_trkparms=pageci%3Afacbe871-2de5-11eb-aa2a-82d017f24296%7Cparentrq%3Af77f924c1750adaa5a919456ffe60390%7Ciid%3A1", "Yeti mic at eBay", true)
	saveDummyLink(database, "BFMV", "https://www.youtube.com/watch?v=-1XOGi-Wc7U", "Bullet For My Valentine live concert video", true)
	saveDummyLink(database, "daxi", "https://www.daxi.re", "Daxi.re", false)
	fmt.Println("Dummy data inserted")

	statistics := stats.New(database, geoIPdb)

	go shortener.StartServer(database, statistics, opts.Host)

	bot.Start(database, statistics, opts.BotToken, opts.Port, opts.AcceptFromUser, opts.Host, opts.IsDebug)
}

func saveDummyLink(database *db.Database, suffix, targetAddress, descr string, isPublic bool) {
	if err := database.SaveShortUrl(suffix, targetAddress, descr, isPublic); err != nil {
		panic(err)
	}
}
