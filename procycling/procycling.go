package procycling

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/ScottFitz/cyclingteammate/model"
)

// BaseURL the procyclingstats base url
const BaseURL string = "https://www.procyclingstats.com"

// SearchURLTemplate the procyclingstats search template
const SearchURLTemplate string = BaseURL + "/search.php?term=%s+%s"

// GotoRiderBaseURL the procyclingstats rider page url
const GotoRiderBaseURL string = BaseURL + "/search.php"

// Search make the ProCycling search request
func Search(firstName, surname string) (*http.Response, error) {
	searchReq, err := http.NewRequest("GET", fmt.Sprintf(SearchURLTemplate, firstName, surname), nil)
	if err != nil {
		log.Printf("could not create search request for '%s', error was: %v", fmt.Sprintf(SearchURLTemplate, firstName, surname), err)
		return nil, fmt.Errorf("could not create search request for '%s', error was: %v", fmt.Sprintf(SearchURLTemplate, firstName, surname), err)
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

// GetRiderPage get the ProCycling rider page
func GetRiderPage(riderLink string) (*http.Response, error) {
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

// ParseMultipleSearchResultsForInividualCyclist parses the multiple results and matches based on name and nationality
func ParseMultipleSearchResultsForInividualCyclist(resultsPage *goquery.Document, firstName, surname, nation string) string {
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
			cyclistLink = GotoRiderBaseURL + link
			return false
		}
		return true
	})
	return cyclistLink
}

// ParseRiderPage parses the ProCyclingStats rider page into the Cyclisy model
func ParseRiderPage(riderPage *goquery.Document, cyclist *model.Cyclist) *model.Cyclist {

	riderPage.Find("body .content div").EachWithBreak(func(index int, item *goquery.Selection) bool {

		return true
	})
	return cyclist
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
