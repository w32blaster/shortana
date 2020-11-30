package geoip

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/cavaliercoder/grab"
	"github.com/oschwald/geoip2-golang"
)

const (
	fileName   = "GeoLite2-City.mmdb"
	tmpTarBall = "/tmp/GeoLite2-City.tar"
)

var m sync.Mutex

type GeoIP struct {
	geoip       *geoip2.Reader
	storagePath string
	licenseKey  string
	isReady     bool
}

func New(path, license string, isReady bool) *GeoIP {

	// TODO: check, if not exists, then download
	geoIPdb, err := geoip2.Open(path + "/" + fileName)
	if err != nil {
		panic(err)
	}

	return &GeoIP{
		geoip:       geoIPdb,
		storagePath: path,
		licenseKey:  license,
		isReady:     isReady,
	}
}

func (g GeoIP) Close() {
	g.geoip.Close()
}

func (g GeoIP) reconnectToDatabase() {
	geoIPdb, err := geoip2.Open(g.storagePath + "/" + fileName)
	if err != nil {
		panic(err)
	}
	g.geoip = geoIPdb
}

// IsReady returns TRUE when a database exists, downloaded, opened and ready to use
func (g GeoIP) IsReady() bool {
	return g.isReady
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

func (g GeoIP) DownloadGeoIPDatabase(fnUpdate func(msg string)) error {

	fnUpdate("start downloading")
	downloadURL := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz", g.licenseKey)
	resp, err := grab.Get("/tmp", downloadURL)
	if err != nil {
		log.Println("Error downloading GeoIP file. Err: " + err.Error())
		return err
	}

	fnUpdate("Downloaded")
	log.Printf("Downloaded file %s with response %d \n", resp.Filename, resp.HTTPResponse.StatusCode)

	folderName := resp.Filename[:len(resp.Filename)-7] // trim the extension ".tar.gz"
	oldDatabase := g.storagePath + "/geocityLite-old.mmdb"

	if err := unGzip(resp.Filename, tmpTarBall); err != nil {
		log.Println("Error while unzipping. Err: " + err.Error())
		return err
	}

	fnUpdate("Downloaded. Untar it")
	if err := untar(tmpTarBall, "/tmp"); err != nil {
		return err
	}

	m.Lock()
	g.isReady = false
	m.Unlock()

	g.Close()

	// rename old database (just in case)
	if err := os.Rename(g.storagePath+"/"+fileName, g.storagePath+"/geocityLite-old.mmdb"); err != nil {
		return err
	}

	// copy freshly downloaded database to our working directory
	fnUpdate("replace database")
	if err := os.Rename(folderName+"/"+fileName, g.storagePath+"/"+fileName); err != nil {
		return err
	}

	g.reconnectToDatabase()

	if err := os.RemoveAll(folderName); err != nil {
		return err
	}
	if err := os.Remove(tmpTarBall); err != nil {
		return err
	}
	if err := os.Remove(resp.Filename); err != nil {
		return err
	}
	if err := os.Remove(oldDatabase); err != nil {
		return err
	}

	m.Lock()
	g.isReady = true
	m.Unlock()

	fnUpdate("Ready")
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
