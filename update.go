package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
)

type geonamesRecord struct {
	ID                  string
	Name, ASCIIName     string
	AlternateNames      []string
	Latitude, Longitude string
	CountryCode         string
	TimezoneID          string
	Population          *big.Int
}

func parseRecord(line string) geonamesRecord {
	tokens := strings.Split(line, string('\t'))

	log.Printf("%+v", tokens)

	// The main 'geoname' table has the following fields :
	// ---------------------------------------------------
	// geonameid         : integer id of record in geonames database
	// name              : name of geographical point (utf8) varchar(200)
	// asciiname         : name of geographical point in plain ascii characters, varchar(200)
	// alternatenames    : alternatenames, comma separated, ascii names automatically transliterated, convenience attribute from alternatename table, varchar(10000)
	// latitude          : latitude in decimal degrees (wgs84)
	// longitude         : longitude in decimal degrees (wgs84)
	// feature class     : see http://www.geonames.org/export/codes.html, char(1)
	// feature code      : see http://www.geonames.org/export/codes.html, varchar(10)
	// country code      : ISO-3166 2-letter country code, 2 characters
	// cc2               : alternate country codes, comma separated, ISO-3166 2-letter country code, 200 characters
	// admin1 code       : fipscode (subject to change to iso code), see exceptions below, see file admin1Codes.txt for display names of this code; varchar(20)
	// admin2 code       : code for the second administrative division, a county in the US, see file admin2Codes.txt; varchar(80)
	// admin3 code       : code for third level administrative division, varchar(20)
	// admin4 code       : code for fourth level administrative division, varchar(20)
	// population        : bigint (8 byte int)
	// elevation         : in meters, integer
	// dem               : digital elevation model, srtm3 or gtopo30, average elevation of 3''x3'' (ca 90mx90m) or 30''x30'' (ca 900mx900m) area in meters, integer. srtm processed by cgiar/ciat.
	// timezone          : the timezone id (see file timeZone.txt) varchar(40)
	// modification date : date of last modification in yyyy-MM-dd format

	pop := new(big.Int)
	pop.SetString(tokens[14], 10)

	return geonamesRecord{
		ID:             tokens[0],
		Name:           tokens[1],
		ASCIIName:      tokens[2],
		AlternateNames: strings.Split(tokens[3], ","),
		Latitude:       tokens[4],
		Longitude:      tokens[5],
		CountryCode:    tokens[8],
		TimezoneID:     tokens[17],
		Population:     pop,
	}
}

type updateCommandInput struct {
	CitySourceURI string
	MinPopulation int64
}

func (u updateCommandInput) ReadCities(f func(r geonamesRecord) error) error {
	log.Printf("Reading cities from %s", u.CitySourceURI)

	var zipr *zip.Reader

	if strings.HasPrefix(u.CitySourceURI, "http") {
		res, err := http.Get(u.CitySourceURI)
		if err != nil {
			return fmt.Errorf("Fetching zip URL %s failed with error %s.", u.CitySourceURI, err)
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("Failed to read response body with error %s", err)
		}

		zipr, err = zip.NewReader(bytes.NewReader(b), int64(len(b)))
	}

	// assume it's a file otherwise
	fr, err := zip.OpenReader(u.CitySourceURI)
	if err != nil {
		return err
	}

	defer fr.Close()
	zipr = &fr.Reader

	for _, file := range zipr.File {
		log.Printf("Reading %s from %s", file.Name, u.CitySourceURI)
		rc, err := file.Open()
		if err != nil {
			return err
		}

		defer rc.Close()

		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			if err := f(parseRecord(scanner.Text())); err != nil {
				return err
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}

	return nil
}

func updateCommand(opts updateCommandInput) {
	db, err := OpenDB()
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	var counter int64

	err = opts.ReadCities(func(record geonamesRecord) error {
		if record.Population.Int64() < opts.MinPopulation {
			return nil
		}

		counter++
		_, err = db.Exec(
			`INSERT OR REPLACE INTO timezone
				(id, name, country, timezone, population)
			VALUES
				(?, ?, ?, ?, ?)`,
			record.ID,
			record.ASCIIName,
			record.CountryCode,
			record.TimezoneID,
			record.Population.String(),
		)

		return err
	})
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	if err := tx.Commit(); err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	fmt.Printf("Updated %d cities\n", counter)
}
