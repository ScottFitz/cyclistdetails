package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// Cyclist details
type Cyclist struct {
	Name        string
	Nationality string
	link        string
}

const proCyclingStatsBaseURL string = "https://www.procyclingstats.com"
const proCyclingStatsSearchURLTemplate string = proCyclingStatsBaseURL + "/search.php?term=%s+%s"

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

	searchReq, err := http.NewRequest("GET", fmt.Sprintf(proCyclingStatsSearchURLTemplate, *firstName, *surname), nil)
	if err != nil {
		log.Printf("could not create view search request for '%s', error was: %v", fmt.Sprintf(proCyclingStatsSearchURLTemplate, *firstName, *surname), err)
		os.Exit(1)
	}

	// Send req using http Client
	client := &http.Client{}
	searchResp, err := client.Do(searchReq)
	if err != nil {
		log.Printf("could not get search results from '%s', error was: %v", fmt.Sprintf(proCyclingStatsSearchURLTemplate, *firstName, *surname), err)
		os.Exit(1)
	}
	defer searchResp.Body.Close()

	if searchResp.StatusCode != http.StatusOK {
		if searchResp.StatusCode == http.StatusTooManyRequests {
			log.Printf("you are being rate limited")
			os.Exit(1)
		}

		log.Printf("bad response from server: %s", searchResp.Status)
		os.Exit(1)
	}

	finalURL := searchResp.Request.URL.String()
	fmt.Printf("The URL request was directed to was: %v\n", finalURL)

	// parse body with goquery.
	doc, err := goquery.NewDocumentFromReader(searchResp.Body)
	if err != nil {
		log.Printf("could not parse page: %v", err)
	}
	cyclist := &Cyclist{}
	if strings.Compare(finalURL, fmt.Sprintf(proCyclingStatsSearchURLTemplate, *firstName, *surname)) == 0 {
		fmt.Printf("There were multiple results for: %s %s\n", *firstName, *surname)

		// extract info we want for each search result, use index and item
		link := parseMultipleSearchResultsForCyclist(doc, *firstName, *surname, *nation)
		log.Printf("Individual link for cyclist: %s", link)
		// Send req to get individual cyclist details
		searchReq, err := http.NewRequest("GET", link, nil)
		if err != nil {
			log.Printf("could not create view search request for '%s', error was: %v", link, err)
			os.Exit(1)
		}
		client := &http.Client{}
		searchResp, err := client.Do(searchReq)
		if err != nil {
			log.Printf("could not get search results from '%s', error was: %v", link, err)
			os.Exit(1)
		}
		defer searchResp.Body.Close()

		if searchResp.StatusCode != http.StatusOK {
			if searchResp.StatusCode == http.StatusTooManyRequests {
				log.Printf("you are being rate limited")
				os.Exit(1)
			}

			log.Printf("bad response from server: %s", searchResp.Status)
			os.Exit(1)
		}

		doc, err := goquery.NewDocumentFromReader(searchResp.Body)
		if err != nil {
			log.Printf("could not parse individual cyclist page: %v", err)
		}

		fmt.Printf("%v\n", doc)
	} else {
		fmt.Printf("Found only one result for: %s %s\n", *firstName, *surname)
	}
	fmt.Printf("\tFound cyclist: %v\n", cyclist)
}

func parseMultipleSearchResultsForCyclist(resultsPage *goquery.Document, firstName, surname, nation string) string {
	var cyclistLink string
	var flag string
	resultsPage.Find("body .content div").EachWithBreak(func(index int, item *goquery.Selection) bool {
		title := item.Text()
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		item.Find("span").Each(func(index int, item *goquery.Selection) {
			spanClass, _ := item.Attr("class")
			if strings.Contains(spanClass, "flags") {
				flag = strings.TrimPrefix(spanClass, "flags ")
			}
		})
		resultName := strings.ToUpper(trimAllSpaces(title))
		searchName := strings.ToUpper(firstName + surname)
		if strings.Contains(resultName, searchName) && flag == nation {
			cyclistLink = proCyclingStatsBaseURL + link
			return false
		}
		return true
	})
	return cyclistLink
}

// simple function to trim all whitespace
func trimAllSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}
