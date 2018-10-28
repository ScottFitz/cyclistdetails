package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ScottFitz/cyclingteammate/procycling"
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

	procyclist, err := procycling.GetCyclistDetails(*firstName, *surname, *nation)
	if err != nil {
		log.Printf("An error occurred getting rider details for '%s - %s' from ProCyclingStats.com, error was: %v", *firstName, *surname, err)
		os.Exit(1)
	}
	fmt.Printf("\tFound cyclist: %v\n", procyclist)
}
