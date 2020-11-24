package db

import (
	"github.com/asdine/storm/v3"
)

type Database struct {
	db *storm.DB
}

func Init(bdb *storm.DB) *Database {
	return &Database{
		db: bdb,
	}
}

func (d Database) SaveShortUrl(shortSuffix, fullTargetAddress string, isPublic bool) error {
	return d.db.Save(&ShortURL{
		ShortUrl: shortSuffix,
		FullUrl:  fullTargetAddress,
		IsPublic: isPublic,
	})
}

func (d Database) GetAll() ([]ShortURL, error) {
	var shortUrls []ShortURL
	err := d.db.All(&shortUrls)
	return shortUrls, err
}
