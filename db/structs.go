package db

import "time"

type (
	ShortURL struct {
		ShortUrl    string `storm:"id"` // unique primary key
		TargetUrl   string
		Description string
		IsPublic    bool
		PublishDate string // format is 2006-01-02
	}

	OneViewStatistic struct {
		ID            int    `storm:"id,increment"`
		UserIpAddress string `storm:"index"`
		ShortUrl      string `storm:"index"` // shortened URL suffix
		Day           string `storm:"index"` // just a date sortable in format of 2020-11-20, to be able to select all the views for a day
		CountryCode   string
		CountryName   string
		City          string
		UserAgent     string
		ViewTimes     []time.Time // one view is one time record, UTC
	}

	// representation only
	OneURLSummaryStatistics struct {
		ShortUrl         string
		PublishDate      string // format is 2006-01-02
		TotalDaysActive  int
		TotalViews       int
		TotalUniqueUsers int
		// TotalCountries  int later
	}
)
