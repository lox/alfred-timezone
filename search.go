package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pascalw/go-alfred"
)

type searchOutputer interface {
	Error(err error)
	Result(r searchResult)
	Flush()
}

type terminalOutputer struct {
}

func (t *terminalOutputer) Error(err error) {
	fmt.Println(err.Error())
}

func (t *terminalOutputer) Result(r searchResult) {
	fmt.Printf("%-25s %-8s %-8s · %-25s · %-20s\n",
		r.Name,
		time.Now().In(r.Location).Format("3:04pm,"),
		time.Now().In(r.Location).Format("Monday"),
		r.CountryName,
		r.Timezone,
	)
}

func (t *terminalOutputer) Flush() {
}

type alfredOutputer struct {
	*alfred.AlfredResponse
}

func newAlfredOutputer() searchOutputer {
	return &alfredOutputer{
		AlfredResponse: alfred.NewResponse(),
	}
}

func (a *alfredOutputer) Error(err error) {
	a.AlfredResponse.AddItem(&alfred.AlfredResponseItem{
		Valid:    false,
		Uid:      "error",
		Title:    "Error Occurred",
		Subtitle: err.Error(),
		Icon:     "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns",
	})
}

func (a *alfredOutputer) Result(r searchResult) {
	descr := fmt.Sprintf("%s, %s", r.Name, r.CountryName)
	t := time.Now().In(r.Location).Format("3:04PM, Monday")

	a.AddItem(&alfred.AlfredResponseItem{
		Valid:    true,
		Uid:      r.ID,
		Title:    fmt.Sprintf("%s: %s", descr, t),
		Subtitle: fmt.Sprintf("%s", r.Location.String()),
		Arg:      fmt.Sprintf("%s: %s", descr, t),
		Icon:     "images/flags-48/" + r.Country + ".png",
	})
}

func (a *alfredOutputer) Flush() {
	a.AlfredResponse.Print()
}

type searchResult struct {
	ID, Name, Country, CountryName, Timezone string
	Location                                 *time.Location
}

func searchTimezones(queryTerms []string) ([]searchResult, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}

	alfred.InitTerms(queryTerms)

	rows, err := db.Query("SELECT id, name, country, timezone FROM timezone")
	if err != nil {
		return nil, err
	}

	results := []searchResult{}
	for rows.Next() {
		var row searchResult
		err = rows.Scan(&row.ID, &row.Name, &row.Country, &row.Timezone)
		if err != nil {
			return results, err
		}

		row.CountryName = lookupCountryName(row.Country)
		matchText := fmt.Sprintf("%s, %s", row.Name, row.CountryName)

		if alfred.MatchesTerms(queryTerms, matchText) {
			l, err := time.LoadLocation(row.Timezone)
			if err != nil {
				return results, err
			}

			row.Location = l
			results = append(results, row)
			log.Printf("HIT %s => %+v", row.ID, row)
		}
	}

	return results, nil
}

func searchCommand(s searchOutputer, queryTerms []string) {
	defer s.Flush()

	results, err := searchTimezones(queryTerms)
	if err != nil {
		s.Error(err)
		os.Exit(1)
	}

	log.Printf("Found %d matches", len(results))

	for _, r := range results {
		s.Result(r)
	}
}
