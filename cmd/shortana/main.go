package main

import (
	"fmt"
	"github.com/w32blaster/shortana/db"
	"time"

	"github.com/w32blaster/shortana/shortener"

	bolt "go.etcd.io/bbolt"
)

func main() {
	fmt.Println("Start the Shortana")

	boltdb, err := bolt.Open("shortana.db", 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		panic(err)
	}

	database, err := db.Init(boltdb)
	if err != nil {
		panic(err)
	}

	shortener.StartServer(database)
}
