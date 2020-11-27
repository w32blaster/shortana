package stats

import (
	"log"
	"net/http"

	"github.com/w32blaster/shortana/db"
	"github.com/w32blaster/shortana/geoip"
)

type Statistics struct {
	db    *db.Database
	geoIP *geoip.GeoIP
}

func New(database *db.Database, geoIPdb *geoip.GeoIP) *Statistics {
	return &Statistics{
		db:    database,
		geoIP: geoIPdb,
	}
}

func (s Statistics) ProcessRequest(req *http.Request, requestedUrl string) {

	ipAddress := req.RemoteAddr
	if len(ipAddress) == 0 {
		ipAddress = req.RemoteAddr
	}

	var countryCode, countryName, city string
	var err error

	if s.geoIP.IsReady() {
		countryCode, countryName, city, err = s.geoIP.GetGeoStatsForTheIP(ipAddress)
		if err != nil {
			log.Println("ERROR! Can't get GeoIP data. Reason: " + err.Error())
		}
	} else {
		log.Println("GeoIP database is not ready yet, so the current view will be saved without GEO data :(")
	}

	err = s.db.SaveStatisticForOneView(ipAddress, requestedUrl, countryCode, countryName, city)
	if err != nil {
		log.Println("ERROR cant save stats because: " + err.Error())
	}
}
