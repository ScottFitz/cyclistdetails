package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ScottFitz/cyclingteammate/model"
	"github.com/ScottFitz/cyclingteammate/procycling"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	var (
		firstName         = flag.String("firstName", "Michael", "First name of cyclist we are generating the report for.")
		surname           = flag.String("surname", "Barry", "Surname of cyclist we are generating the report for.")
		nation            = flag.String("nation", "ca", "Two letter nation code, e.g. ca, es, ie, etc.")
		teammateFirstName = flag.String("teammateFirstName", "Antonio", "First name of possible team mate.")
		teammateSurname   = flag.String("teammateSurame", "Flescha", "Surname of possible team mate.")
	)
	flag.Parse()

	// validate mandatory flags are supplied
	if *firstName == "" || *surname == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *teammateFirstName != "" && *teammateSurname == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *teammateFirstName == "" && *teammateSurname != "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	searchResp, err := procycling.Search(*firstName, *surname)
	if err != nil {
		log.Printf("could not get search results for '%s - %s', error was: %v", *firstName, *surname, err)
		os.Exit(1)
	}
	defer searchResp.Body.Close()

	finalURL := searchResp.Request.URL.String()
	fmt.Printf("The URL request was directed to was: %v\n", finalURL)

	// parse body with goquery.
	riderPage, err := goquery.NewDocumentFromReader(searchResp.Body)
	if err != nil {
		log.Printf("could not parse page: %v", err)
	}

	cyclist := &model.Cyclist{}
	cyclist.Link = searchResp.Request.URL.String()
	cyclist.Name = *firstName + " " + *surname
	cyclist.Nationality = *nation
	// if more than one serch result, parse out specific rider and get that rider page
	if strings.Compare(finalURL, fmt.Sprintf(procycling.SearchURLTemplate, *firstName, *surname)) == 0 {
		fmt.Printf("There were multiple results for: %s %s\n", *firstName, *surname)

		link := procycling.ParseMultipleSearchResultsForInividualCyclist(riderPage, *firstName, *surname, *nation)
		log.Printf("Individual link for cyclist: %s", link)

		resp, err := procycling.GetRiderPage(link)
		if err != nil {
			log.Printf("could not get rider page for '%s', error was: %v", link, err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		cyclist.Link = resp.Request.URL.String()

		riderPage, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Printf("could not parse individual cyclist page: %v", err)
		}

		fmt.Printf("%v\n", riderPage)
	}

	fmt.Printf("\tFound cyclist: %v\n", cyclist)
	procycling.ParseRiderPage(riderPage, cyclist)
}
