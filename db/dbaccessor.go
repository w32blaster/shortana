package db

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/msgpack"
	"github.com/asdine/storm/v3/q"
	"go.etcd.io/bbolt"
	"strings"
	"time"
)

const (
	DayFormat = "2006-01-02"
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
	day := time.Now().UTC().Format(DayFormat)

	return d.db.Save(&ShortURL{
		ShortUrl:    shortSuffix,
		TargetUrl:   fullTargetAddress,
		Description: description,
		IsPublic:    isPublic,
		PublishDate: day,
	})
}

func (d Database) SaveShortUrlObject(shortUrl *ShortURL) error {
	shortUrl.PublishDate = time.Now().UTC().Format(DayFormat)
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

func (d Database) IsEmpty() bool {
	var shortUrls []ShortURL
	d.db.All(&shortUrls)
	return len(shortUrls) == 0
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

func (d Database) GetUrlByID(ID int) (*ShortURL, error) {
	var shortUrl ShortURL
	err := d.db.One("ID", ID, &shortUrl)
	return &shortUrl, err
}

func (d Database) GetStatisticsForOneURL(shortUrlID int) (*ShortURL, map[string]OneDaySummaryStatistics, error) {

	sURL, err := d.GetUrlByID(shortUrlID)
	if err != nil {
		return nil, nil, err
	}

	var views []OneViewStatistic
	if err := d.db.Find("ShortUrl", sURL.ShortUrl, &views); err != nil {
		return nil, nil, err
	}

	mapViews := make(map[string]OneDaySummaryStatistics)
	for _, k := range views {
		if existingView, found := mapViews[k.Day]; found {
			existingView.TotalViews = existingView.TotalViews + len(k.ViewTimes)
			existingView.UniqueViews++
			mapViews[k.Day] = existingView

		} else {
			mapViews[k.Day] = OneDaySummaryStatistics{
				Date:               k.Day,
				DateWithoutHyphens: strings.ReplaceAll(k.Day, "-", ""),
				TotalViews:         len(k.ViewTimes),
				UniqueViews:        1,
			}
		}
	}

	return sURL, mapViews, nil
}

func (d Database) GetStatisticForOneURLOneDay(shortUrlID int, dayDate time.Time) (*ShortURL, []OneViewStatistic, error) {

	sURL, err := d.GetUrlByID(shortUrlID)
	if err != nil {
		return nil, nil, err
	}

	query := d.db.Select(
		q.And(
			q.Eq("ShortUrl", sURL.ShortUrl),
			q.Eq("Day", dayDate.Format(DayFormat)),
		),
	)

	var foundViews []OneViewStatistic
	err = query.Find(&foundViews)
	return sURL, foundViews, err
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
			pDate, _ := time.Parse(DayFormat, publishDate)
			duration := time.Now().Sub(pDate)

			// add a new record to the map
			groupedStats[k.ShortUrl] = OneURLSummaryStatistics{
				ID:               k.ID,
				ShortUrlID:       allShortUrls[k.ShortUrl].ID,
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
	day := now.Format(DayFormat)

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

func (d Database) GetViewByID(ID int) (*OneViewStatistic, error) {
	var view OneViewStatistic
	err := d.db.One("ID", ID, &view)
	return &view, err
}
