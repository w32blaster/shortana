package main

import (
	"fmt"
	"time"

	"github.com/w32blaster/shortana/db"
	"github.com/w32blaster/shortana/shortener"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/msgpack"
	"go.etcd.io/bbolt"
)

func main() {
	fmt.Println("Start the Shortana")

	// config:
	// - hostname

	boltdb, err := storm.Open("shortana.db", storm.Codec(msgpack.Codec), storm.BoltOptions(0600, &bbolt.Options{Timeout: 5 * time.Second}))
	if err != nil {
		panic(err)
	}

	database := db.Init(boltdb)

	saveDummyLink(database, "STORM", "https://github.com/asdine/storm#options", true)
	saveDummyLink(database, "yeti", "https://www.ebay.co.uk/itm/Blue-Yeti-Professional-Multi-Pattern-USB-Mic-for-Record-and-Stream-Sunset-Sky/274582563719?_trkparms=aid%3D777001%26algo%3DDISCO.FEED%26ao%3D1%26asc%3D20200211172457%26meid%3Dd44c41c9188544f0b2a06ab469877571%26pid%3D101213%26rk%3D1%26rkt%3D1%26mehot%3Dnone%26itm%3D274582563719%26pmt%3D0%26noa%3D1%26pg%3D2380057%26algv%3DRecommendingSearch%26brand%3DBlue+Microphones&_trksid=p2380057.c101213.m46344&_trkparms=pageci%3Afacbe871-2de5-11eb-aa2a-82d017f24296%7Cparentrq%3Af77f924c1750adaa5a919456ffe60390%7Ciid%3A1", true)
	saveDummyLink(database, "BFMV", "https://www.youtube.com/watch?v=-1XOGi-Wc7U", true)
	saveDummyLink(database, "daxi", "https://www.daxi.re", false)
	fmt.Println("Dummy data inserted")

	shortener.StartServer(database)
}

func saveDummyLink(database *db.Database, suffix, targetAddress string, isPublic bool) {
	if err := database.SaveShortUrl(suffix, targetAddress, isPublic); err != nil {
		panic(err)
	}
}
