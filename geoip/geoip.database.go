package geoip

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/cavaliercoder/grab"
	"github.com/oschwald/geoip2-golang"
)

const (
	fileName = "GeoLite2-City.mmdb"
)

type GeoIP struct {
	geoip      *geoip2.Reader
	geoipPath  string
	licenseKey string
}

func New(path, license string) *GeoIP {

	// TODO: check, if not exists, then download
	geoIPdb, err := geoip2.Open(path + "/" + fileName)
	if err != nil {
		panic(err)
	}

	return &GeoIP{
		geoip:      geoIPdb,
		geoipPath:  path,
		licenseKey: license,
	}
}

func (g GeoIP) Close() {
	g.geoip.Close()
}

func (g GeoIP) reconnectToDatabase() {
	g.geoip.Close()
	geoIPdb, err := geoip2.Open(g.geoipPath + "/" + fileName)
	if err != nil {
		panic(err)
	}
	g.geoip = geoIPdb
}

func (g GeoIP) GetGeoStatsForTheIP(ipAddress string) (string, string, string, error) {
	if len(ipAddress) == 0 {
		return "unknown", "unknown", "unknown", errors.New("IP Address of visitor is unknown")
	}

	ip := net.ParseIP(ipAddress)
	record, err := g.geoip.City(ip)
	if err != nil {
		return "unknown", "unknown", "unknown", err
	}

	return record.Country.IsoCode, record.Country.Names["en"], record.City.Names["en"], nil
}

func (g GeoIP) DownloadGeoIPDatabase() error {

	downloadURL := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz", g.licenseKey)
	resp, err := grab.Get(g.geoipPath, downloadURL)
	if err != nil {
		return err
	}

	if err := unGzip(g.geoipPath+"/"+resp.Filename, g.geoipPath+"/geocityLite.tar"); err != nil {
		return err
	}

	err = os.Rename(g.geoipPath+"/"+fileName, g.geoipPath+"/geocityLite-old.mmdb")
	if err != nil {
		return err
	}

	if err := untar(g.geoipPath+"/geocityLite.tar", g.geoipPath+"/"+fileName); err != nil {
		return err
	}

	g.reconnectToDatabase()

	return nil
}

func untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func unGzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}
