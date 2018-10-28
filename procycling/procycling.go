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

const baseURL string = "https://www.procyclingstats.com"
const searchURLTemplate string = baseURL + "/search.php?term=%s+%s"
const gotoRiderBaseURL string = baseURL + "/search.php"

// GetCyclistDetails get the full cyclist details from ProCyclingStats
func GetCyclistDetails(firstName, surname, nation string) (*model.Cyclist, error) {
	searchResp, err := search(firstName, surname)
	if err != nil {
		log.Printf("could not get search results for '%s - %s', error was: %v", firstName, surname, err)
		return nil, fmt.Errorf("could not get search results for '%s - %s', error was: %v", firstName, surname, err)
	}
	defer searchResp.Body.Close()

	finalURL := searchResp.Request.URL.String()
	fmt.Printf("The URL request was directed to was: %v\n", finalURL)

	// parse body with goquery.
	riderPage, err := goquery.NewDocumentFromReader(searchResp.Body)
	if err != nil {
		log.Printf("could not parse page: %v", err)
		return nil, fmt.Errorf("could not parse page: %v", err)
	}

	cyclist := &model.Cyclist{}
	cyclist.Link = searchResp.Request.URL.String()
	cyclist.Name = firstName + " " + surname
	cyclist.Nationality = nation
	// if more than one search result, parse out specific rider and get that rider page
	if strings.Compare(finalURL, fmt.Sprintf(searchURLTemplate, firstName, surname)) == 0 {
		fmt.Printf("There were multiple results for: %s %s\n", firstName, surname)

		link := parseMultipleSearchResultsForInividualCyclist(riderPage, firstName, surname, nation)
		log.Printf("Individual link for cyclist: %s", link)

		resp, err := getRiderPage(link)
		if err != nil {
			log.Printf("could not get rider page for '%s', error was: %v", link, err)
			return nil, fmt.Errorf("could not get rider page for '%s', error was: %v", link, err)
		}
		defer resp.Body.Close()

		cyclist.Link = resp.Request.URL.String()

		riderPage, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Printf("could not parse individual cyclist page: %v", err)
			return nil, fmt.Errorf("could not parse individual cyclist page: %v", err)
		}

		fmt.Printf("%v\n", riderPage)
		parseRiderPage(riderPage, cyclist)
	}
	return cyclist, nil
}

// Search: make the ProCycling search request
func search(firstName, surname string) (*http.Response, error) {
	searchReq, err := http.NewRequest("GET", fmt.Sprintf(searchURLTemplate, firstName, surname), nil)
	if err != nil {
		log.Printf("could not create search request for '%s', error was: %v", fmt.Sprintf(searchURLTemplate, firstName, surname), err)
		return nil, fmt.Errorf("could not create search request for '%s', error was: %v", fmt.Sprintf(searchURLTemplate, firstName, surname), err)
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

// GetRiderPage: get the ProCycling rider page
func getRiderPage(riderLink string) (*http.Response, error) {
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

// ParseMultipleSearchResultsForInividualCyclist: parses the multiple results and matches based on name and nationality
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
			cyclistLink = gotoRiderBaseURL + link
			return false
		}
		return true
	})
	return cyclistLink
}

// ParseRiderPage: parses the ProCyclingStats rider page into the Cyclisy model
func parseRiderPage(riderPage *goquery.Document, cyclist *model.Cyclist) *model.Cyclist {

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
