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

	countryCode, countryName, city, err := s.geoIP.GetGeoStatsForTheIP(ipAddress)
	if err != nil {
		log.Println("ERROR! Can't get GeoIP data. Reason: " + err.Error())
	}

	err = s.db.SaveStatisticForOneView(ipAddress, requestedUrl, countryCode, countryName, city)
	if err != nil {
		log.Println("ERROR cant save stats because: " + err.Error())
	}
}
