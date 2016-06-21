package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"

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
		var out searchOutputer

		if *searchAlfred {
			out = newAlfredOutputer()
		} else {
			out = &terminalOutputer{}
		}

		searchCommand(out, *searchFilters)
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
