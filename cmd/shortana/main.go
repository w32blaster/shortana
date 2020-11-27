package main

import (
	"fmt"
	"github.com/w32blaster/shortana/bot"
	"github.com/w32blaster/shortana/db"
	"github.com/w32blaster/shortana/geoip"
	"github.com/w32blaster/shortana/shortener"
	"github.com/w32blaster/shortana/stats"

	"github.com/caarlos0/env"
)

type Opts struct {
	Port              int    `env:"PORT" envDefault:"8444"`
	Host              string `env:"HOST" envDefault:"http://localhost:3000"`
	IsDebug           bool   `env:"IS_DEBUG"`
	BotToken          string `env:"BOT_TOKEN,required"`
	AcceptFromUser    int    `env:"ACCEPT_FROM_USER"`
	GeoIpDatabasePath string `env:"GEOIP_DB_PATH" envDefault:"."`
	MaxmindLicenseKey string `env:"MAXMIND_LICENSE_KEY,required"`
}

func main() {
	fmt.Println("Start the Shortana")

	// parse flags
	var opts = Opts{}
	if err := env.Parse(&opts); err != nil {
		panic("Can't parse ENV VARS: " + err.Error())
	}

	// open the GeoIP database
	geoIP := geoip.New(opts.GeoIpDatabasePath, opts.MaxmindLicenseKey)
	defer geoIP.Close()

	// Init BoltDB database
	database := db.Init()
	defer database.Close()

	statistics := stats.New(database, geoIP)

	// for development only
	saveDummyLink(database, "STORM", "https://github.com/asdine/storm#options", "Storm project at GitHub", true)
	saveDummyLink(database, "yeti", "https://www.ebay.co.uk/itm/Blue-Yeti-Professional-Multi-Pattern-USB-Mic-for-Record-and-Stream-Sunset-Sky/274582563719?_trkparms=aid%3D777001%26algo%3DDISCO.FEED%26ao%3D1%26asc%3D20200211172457%26meid%3Dd44c41c9188544f0b2a06ab469877571%26pid%3D101213%26rk%3D1%26rkt%3D1%26mehot%3Dnone%26itm%3D274582563719%26pmt%3D0%26noa%3D1%26pg%3D2380057%26algv%3DRecommendingSearch%26brand%3DBlue+Microphones&_trksid=p2380057.c101213.m46344&_trkparms=pageci%3Afacbe871-2de5-11eb-aa2a-82d017f24296%7Cparentrq%3Af77f924c1750adaa5a919456ffe60390%7Ciid%3A1", "Yeti mic at eBay", true)
	saveDummyLink(database, "BFMV", "https://www.youtube.com/watch?v=-1XOGi-Wc7U", "Bullet For My Valentine live concert video", true)
	saveDummyLink(database, "daxi", "https://www.daxi.re", "Daxi.re", false)
	fmt.Println("Dummy data inserted")

	// Run web server
	go shortener.StartServer(database, statistics, opts.Host)

	// Run Telegram bot
	bot.Start(database, statistics, geoIP, opts.BotToken, opts.Port, opts.AcceptFromUser, opts.Host, opts.IsDebug)
}

func saveDummyLink(database *db.Database, suffix, targetAddress, descr string, isPublic bool) {
	if err := database.SaveShortUrl(suffix, targetAddress, descr, isPublic); err != nil {
		panic(err)
	}
}
