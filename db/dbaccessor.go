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

func Init(storagePath string) *Database {

	// Open Storm DB
	boltdb, err := storm.Open(storagePath+"/shortana.db", storm.Codec(msgpack.Codec), storm.BoltOptions(0600, &bbolt.Options{Timeout: 5 * time.Second}))
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
	day := time.Now().UTC().Format(dayFormat)

	return d.db.Save(&ShortURL{
		ShortUrl:    shortSuffix,
		TargetUrl:   fullTargetAddress,
		Description: description,
		IsPublic:    isPublic,
		PublishDate: day,
	})
}

func (d Database) SaveShortUrlObject(shortUrl *ShortURL) error {
	shortUrl.PublishDate = time.Now().UTC().Format(dayFormat)
	return d.db.Save(shortUrl)
}

func (d Database) UpdateShortUrl(id, fieldName string, value interface{}) error {
	shortUrl, err := d.GetUrl(id)
	if err != nil {
		return err
	}
	return d.db.UpdateField(shortUrl, fieldName, value)
}

func (d Database) GetAll() ([]ShortURL, error) {
	var shortUrls []ShortURL
	err := d.db.All(&shortUrls)
	return shortUrls, err
}

func (d Database) GetAllMapped() (map[string]ShortURL, error) {
	var shortUrls []ShortURL
	err := d.db.All(&shortUrls)

	mapped := make(map[string]ShortURL)
	for _, k := range shortUrls {
		mapped[k.ShortUrl] = k
	}
	return mapped, err
}

func (d Database) GetUrl(suffix string) (*ShortURL, error) {
	var shortUrl ShortURL
	err := d.db.One("ShortUrl", suffix, &shortUrl)
	return &shortUrl, err
}

func (d Database) GetAllStatisticsGroupedByURLs() (map[string]OneURLSummaryStatistics, error) {
	groupedStats := make(map[string]OneURLSummaryStatistics)

	var stats []OneViewStatistic
	err := d.db.All(&stats)
	if err != nil {
		return groupedStats, err
	}

	allShortUrls, err := d.GetAllMapped()
	if err != nil {
		return groupedStats, err
	}

	for _, k := range stats {
		if oneShortUrl, found := groupedStats[k.ShortUrl]; found {

			// update record that is already in the map
			oneShortUrl.TotalViews = oneShortUrl.TotalViews + len(k.ViewTimes)
			oneShortUrl.TotalUniqueUsers++
			groupedStats[k.ShortUrl] = oneShortUrl

		} else {

			publishDate := allShortUrls[k.ShortUrl].PublishDate
			pDate, _ := time.Parse(dayFormat, publishDate)
			duration := time.Now().Sub(pDate)

			// add a new record to the map
			groupedStats[k.ShortUrl] = OneURLSummaryStatistics{
				ShortUrl:         k.ShortUrl,
				PublishDate:      publishDate,
				TotalDaysActive:  int(duration.Hours() / 24),
				TotalViews:       len(k.ViewTimes),
				TotalUniqueUsers: 1,
			}
		}
	}

	return groupedStats, err
}

func (d Database) SaveStatisticForOneView(ipAddress, requestedUrl, countryCode, countryName, city, userAgent string) error {

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
			UserAgent:     userAgent,
			ViewTimes:     []time.Time{now},
		})
	}

	// update existing one
	foundView.ViewTimes = append(foundView.ViewTimes, now)
	return d.db.Update(&foundView)
}
