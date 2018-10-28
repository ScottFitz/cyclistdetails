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
const proCyclingStatsGotoRiderBaseURL string = proCyclingStatsBaseURL + "/search.php"

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

	searchResp, err := makeProCyclingSearch(*firstName, *surname)
	if err != nil {
		log.Printf("could not get search results for '%s - %s', error was: %v", *firstName, *surname, err)
		os.Exit(1)
	}
	defer searchResp.Body.Close()

	finalURL := searchResp.Request.URL.String()
	fmt.Printf("The URL request was directed to was: %v\n", finalURL)

	// parse body with goquery.
	resultsPage, err := goquery.NewDocumentFromReader(searchResp.Body)
	if err != nil {
		log.Printf("could not parse page: %v", err)
	}
	cyclist := &Cyclist{}
	if strings.Compare(finalURL, fmt.Sprintf(proCyclingStatsSearchURLTemplate, *firstName, *surname)) == 0 {
		fmt.Printf("There were multiple results for: %s %s\n", *firstName, *surname)

		// extract info we want for each search result, use index and item
		link := parseMultipleSearchResultsForInividualCyclist(resultsPage, *firstName, *surname, *nation)
		log.Printf("Individual link for cyclist: %s", link)

		resp, err := getProCyclingRiderPage(link)
		if err != nil {
			log.Printf("could not get rider page for '%s', error was: %v", link, err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		riderPage, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Printf("could not parse individual cyclist page: %v", err)
		}

		fmt.Printf("%v\n", riderPage)
	} else {
		fmt.Printf("Found only one result for: %s %s\n", *firstName, *surname)
	}
	fmt.Printf("\tFound cyclist: %v\n", cyclist)
}

func makeProCyclingSearch(firstName, surname string) (*http.Response, error) {
	searchReq, err := http.NewRequest("GET", fmt.Sprintf(proCyclingStatsSearchURLTemplate, firstName, surname), nil)
	if err != nil {
		log.Printf("could not create search request for '%s', error was: %v", fmt.Sprintf(proCyclingStatsSearchURLTemplate, firstName, surname), err)
		return nil, fmt.Errorf("could not create search request for '%s', error was: %v", fmt.Sprintf(proCyclingStatsSearchURLTemplate, firstName, surname), err)
	}
	client := &http.Client{}
	searchResp, err := client.Do(searchReq)
	if err != nil {
		log.Printf("could not get search results from '%s', error was: %v", searchReq.URL, err)
		return nil, fmt.Errorf("bad response from server: %s", err)
	}

	if searchResp.StatusCode != http.StatusOK {
		if searchResp.StatusCode == http.StatusTooManyRequests {
			log.Printf("you are being rate limited: %s", err)
			return nil, fmt.Errorf("you are being rate limited: %s", err)
		}

		log.Printf("bad response from server: %s", searchResp.Status)
		return nil, fmt.Errorf("bad response from server: %s", err)
	}
	return searchResp, nil
}

func getProCyclingRiderPage(riderLink string) (*http.Response, error) {
	// Send req to get individual cyclist details
	req, err := http.NewRequest("GET", riderLink, nil)
	if err != nil {
		log.Printf("could not create ger rider page request for '%s', error was: %v", riderLink, err)
		return nil, fmt.Errorf("could not create ger rider page request for '%s', error was: %v", riderLink, err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("could not get rider page results from '%s', error was: %v", riderLink, err)
		return nil, fmt.Errorf("could not get rider page results from '%s', error was: %v", riderLink, err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			log.Printf("you are being rate limited: %s", err)
			return nil, fmt.Errorf("you are being rate limited: %s", err)
		}

		log.Printf("bad response from server: %s", resp.Status)
		return nil, fmt.Errorf("bad response from server: %s", err)
	}

	return resp, nil
}

func parseMultipleSearchResultsForInividualCyclist(resultsPage *goquery.Document, firstName, surname, nation string) string {
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
			cyclistLink = proCyclingStatsGotoRiderBaseURL + link
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
