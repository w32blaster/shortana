package db

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/msgpack"
	"github.com/asdine/storm/v3/q"
	"go.etcd.io/bbolt"
	"time"
)

const (
	dayFormat = "2006-01-02"
)

type Database struct {
	db           *storm.DB
	licenseKey   string
	databasePath string
}

func Init() *Database {

	// Open Storm DB
	boltdb, err := storm.Open("shortana.db", storm.Codec(msgpack.Codec), storm.BoltOptions(0600, &bbolt.Options{Timeout: 5 * time.Second}))
	if err != nil {
		panic(err)
	}

	return &Database{
		db: boltdb,
	}
}

func (d Database) Close() {
	d.db.Close()
}

func (d Database) SaveShortUrl(shortSuffix, fullTargetAddress, description string, isPublic bool) error {
	return d.db.Save(&ShortURL{
		ShortUrl:    shortSuffix,
		FullUrl:     fullTargetAddress,
		Description: description,
		IsPublic:    isPublic,
	})
}

func (d Database) GetAll() ([]ShortURL, error) {
	var shortUrls []ShortURL
	err := d.db.All(&shortUrls)
	return shortUrls, err
}

func (d Database) GetUrl(suffix string) (ShortURL, error) {
	var shortUrl ShortURL
	err := d.db.One("ShortUrl", suffix, &shortUrl)
	return shortUrl, err
}

func (d Database) GetAllStatistics() ([]OneViewStatistic, error) {
	var stats []OneViewStatistic
	err := d.db.All(&stats)
	return stats, err
}

func (d Database) SaveStatisticForOneView(ipAddress, requestedUrl, countryCode, countryName, city string) error {

	now := time.Now().UTC()
	day := now.Format(dayFormat)

	// firstly, find whether this user has already accessed this URL
	query := d.db.Select(
		q.And(
			q.Eq("UserIpAddress", ipAddress),
			q.Eq("ShortUrl", requestedUrl),
			q.Eq("Day", day),
		),
	)

	var foundView OneViewStatistic
	err := query.First(&foundView)
	if err != nil {

		// not found, create a fresh record
		return d.db.Save(&OneViewStatistic{
			UserIpAddress: ipAddress,
			ShortUrl:      requestedUrl,
			CountryCode:   countryCode,
			CountryName:   countryName,
			City:          city,
			Day:           day,
			ViewTimes:     []time.Time{now},
		})
	}

	// update existing one
	foundView.ViewTimes = append(foundView.ViewTimes, now)
	return d.db.Update(&foundView)
}
