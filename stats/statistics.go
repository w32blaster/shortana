package stats

import (
	"errors"
	"github.com/oschwald/geoip2-golang"
	"github.com/w32blaster/shortana/db"
	"log"
	"net"
	"net/http"
)

type Statistics struct {
	db    *db.Database
	geoip *geoip2.Reader
}

func New(database *db.Database, geoIPdb *geoip2.Reader) *Statistics {
	return &Statistics{
		db:    database,
		geoip: geoIPdb,
	}
}

func (s Statistics) ProcessRequest(req *http.Request, requestedUrl string) {

	ipAddress := req.RemoteAddr
	if len(ipAddress) == 0 {
		ipAddress = req.RemoteAddr
	}

	countryCode, countryName, city, err := s.getGeoStatsForTheIP(ipAddress)
	if err != nil {
		log.Println("ERROR! Can't get GeoIP data. Reason: " + err.Error())
	}

	err = s.db.SaveStatisticForOneView(ipAddress, requestedUrl, countryCode, countryName, city)
	if err != nil {
		log.Println("ERROR cant save stats because: " + err.Error())
	}
}

func (s Statistics) getGeoStatsForTheIP(ipAddress string) (string, string, string, error) {
	if len(ipAddress) == 0 {
		return "unknown", "unknown", "unknown", errors.New("IP Address of visitor is unknown")
	}

	ip := net.ParseIP(ipAddress)
	record, err := s.geoip.City(ip)
	if err != nil {
		return "unknown", "unknown", "unknown", err
	}

	return record.Country.IsoCode, record.Country.Names["en"], record.City.Names["en"], nil
}
