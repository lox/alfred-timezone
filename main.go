package main

import (
	"io/ioutil"
	"log"
	"os"

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
		searchAlfred        = search.Flag("alfred", "Results as Alfred XML rather than a list").Bool()
		update              = kingpin.Command("update", "Updates repositories from geonames")
		updateSource        = update.Flag("source", "Either a file or a url of a geonames dump file").Default(defaultGeonamesDump).String()
		updateMinPopulation = update.Flag("min-population", "A minimum population to require for cities").Default("50000").Int64()
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
		var out searchOutputer

		if *searchAlfred {
			out = newAlfredOutputer()
		} else {
			out = &terminalOutputer{}
		}

		searchCommand(out, *searchFilters)
	}
}
