package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/pascalw/go-alfred"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	defaultGeonamesDump = "http://download.geonames.org/export/dump/cities15000.zip"
)

func main() {
	var (
		debug               = kingpin.Flag("debug", "Show debugging output").Bool()
		search              = kingpin.Command("search", "Search timezones")
		searchFilters       = search.Arg("filter", "Filter strings").Strings()
		update              = kingpin.Command("update", "Updates repositories from Github")
		updateSource        = update.Flag("source", "Either a file or a url of a geonames dump file").Default(defaultGeonamesDump).String()
		updateMinPopulation = update.Flag("min-population", "A minimum population to require for cities").Int64()
	)

	cmd := kingpin.Parse()

	if *debug == false {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}

	switch cmd {
	case update.FullCommand():
		updateCommand(updateCommandInput{
			CitySourceURI: *updateSource,
			MinPopulation: *updateMinPopulation,
		})
	case search.FullCommand():
		timezonesCommand(*searchFilters)
	}
}

func alfredError(err error) *alfred.AlfredResponseItem {
	return &alfred.AlfredResponseItem{
		Valid:    false,
		Uid:      "error",
		Title:    "Error Occurred",
		Subtitle: err.Error(),
		Icon:     "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns",
	}
}

func backgroundUpdate() error {
	cmd := exec.Command(os.Args[0], "update")
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Printf("Background pid %#v", cmd.Process.Pid)
	return nil
}

func timezonesCommand(queryTerms []string) {
	alfred.InitTerms(queryTerms)

	response := alfred.NewResponse()
	defer response.Print()

	db, err := OpenDB()
	if err != nil {
		response.AddItem(alfredError(err))
		return
	}

	rows, err := db.Query("SELECT id, name, country, timezone FROM timezone")
	if err != nil {
		response.AddItem(alfredError(err))
		return
	}

	for rows.Next() {
		var id, name, country, timezone string
		err = rows.Scan(&id, &name, &country, &timezone)
		if err != nil {
			response.AddItem(alfredError(err))
			return
		}

		descr := fmt.Sprintf("%s, %s", name, lookupCountryName(country))

		if alfred.MatchesTerms(queryTerms, descr) {
			l, err := time.LoadLocation(timezone)
			if err != nil {
				response.AddItem(alfredError(err))
				return
			}

			t := time.Now().In(l).Format("Monday, 3:04PM")

			response.AddItem(&alfred.AlfredResponseItem{
				Valid:    true,
				Uid:      id,
				Title:    fmt.Sprintf("%s 🕐 %s", descr, t),
				Subtitle: fmt.Sprintf("%s", l.String()),
				Arg:      fmt.Sprintf("%s 🕐 %s", descr, t),
				Icon:     "images/flags-48/" + country + ".png",
			})
		}
	}
}
