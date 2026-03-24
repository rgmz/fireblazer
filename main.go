package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"slices"
	"strings"

	utils "github.com/bedros-p/fireblazer/utils"

	"github.com/yarlson/pin"
)

var key = flag.String("apiKey", "", "API key to scan. Can also be your first positional arg.")
var referrer = flag.String("referrer", "", "Set the referrer (Referer:) header for when an API key is restricted to a website.")
var dangerouslySkipVerification = flag.Bool("dangerouslySkipVerification", false, "Skip API key verification")
var workerCount = flag.Int("workerCount", 170, "Set the amount of worker threads to spawn for executing the requests")

// interactive|text|json|yaml
var outputFormat = flag.String("outputFormat", "interactive", "[WIP] Output format (interactive|text|json|yaml)")
var outputDetails = flag.String("outputDetails", "full", "[WIP] Comma delimited list of what to include in the details (description|title|name). Comma delimited.")
var isInteractive = *outputFormat == "interactive" || *outputFormat == ""
var timingEnabled = flag.Bool("findSlowService", false, "[DEBUG] Find which service took the longest to test + elapsed time. Use to file an issue for program hangs.")

type APIDetails struct {
	Description string
	Title       string
}

func processKey(k string, gapiServices []utils.Service, blacklisted []string, falsePos []string) {
	if *dangerouslySkipVerification {
		if isInteractive || *outputFormat == "text" {
			log.Println("Skipping API key verification.")
		}
	} else if valid, err := utils.TestKeyValidity(k); !valid {
		if err != nil {
			log.Printf("Error testing API key validity for %s: %v\n. Ensure that you can connect to https://generativelanguage.googleapis.com as it's used for checking key validity. To skip primary validation (at risk of invalid results), use the -dangerouslySkipVerification flag.", k, err)
			return
		}

		log.Printf("Invalid API key: %s\n", k)
		log.Println("If you're sure the key is valid, use the -dangerouslySkipVerification flag [fireblazer -dangerouslySkipVerification AIza-KeYHere]")
		return
	} else {
		if isInteractive || *outputFormat == "text" {
			log.Println("Valid API key, proceeding.")
		}
	}

	var scanPin *pin.Pin
	var cancel context.CancelFunc
	var updateCh chan utils.ScanUpdate
	var updateDone chan struct{}

	if isInteractive {
		scanPin = pin.New("Scanning...")
		cancel = scanPin.Start(context.Background())
		defer cancel()

		updateCh = make(chan utils.ScanUpdate, *workerCount)
		updateDone = make(chan struct{})

		go func() {
			for update := range updateCh {
				scanPin.UpdateMessage(fmt.Sprintf("Service count - %d in scope. Scanning %d more... %v", update.CurrentFound, update.CurrentRem, update.ItemCleanName))
			}
			close(updateDone)
		}()
	}

	foundServices, failCount, maxTime := utils.ScanServices(k, *referrer, gapiServices, blacklisted, falsePos, *workerCount, *timingEnabled, updateCh)

	if isInteractive {
		<-updateDone
		scanPin.Stop(fmt.Sprintf("Scan complete! Identified %d services available in the project.", len(foundServices)))
	} else {
		log.Printf("Scan complete! Identified %d services available in the project.", len(foundServices))
	}

	log.Println("APIs available to this API key:")

	for _, service := range foundServices {
		if slices.Contains(falsePos, service) {
			// Commented out - I only need to have them here as a reminder, dw, just so i know i should work on those.
			// log.Printf(" - %s.googleapis.com (false positive)", service)
		} else {
			log.Printf(" - %s.googleapis.com", service)
		}
	}

	log.Printf("All discovery endpoint tests completed with %d failures.", failCount)

	if *timingEnabled {
		log.Printf("Longest running service - %v\n\n\n", maxTime)
	}
}

func main() {
	flag.Parse()

	// utils.MultipartAllDiscoveries(*key, []string{"generativelanguage.googleapis.com", "discovery.googleapis.com"})
	// return

	falsePos := []string{
		"digitalassetlinks",
		"oauth2",
		"servicecontrol",
		"storage",
	}

	//  those don't work / hang the program - all that hang are deprecated anyways, so it's blank for now
	blacklisted := []string{}

	gapiServices := make([]utils.Service, 0)

	for _, raw := range utils.GoogleApiList {
		hostname := strings.Split(raw, "/")[0]
		cleanName := strings.Split(hostname, ".")[0]
		discoveryUrl := "https://" + hostname + "/$discovery/rest"

		gapiServices = append(gapiServices, utils.Service{
			CleanName:    cleanName,
			DiscoveryUrl: discoveryUrl,
		})
	}

	keys := []string{}
	if *key != "" {
		keys = append(keys, *key)
	}
	keys = append(keys, flag.Args()...)

	if len(keys) == 0 {
		log.Fatal("You must provide at least one API key. You can pass it as a named flag or as positional arguments. Usage samples: \n - \"fireblazer AIza-key1 AIza-key2\" \n - \"fireblazer -apiKey AIza-key\". \nTerminating.")
	}

	if isInteractive || *outputFormat == "text" {
		log.Printf("Successfully loaded %d discovery endpoints from hardcoded list.", len(gapiServices))
	}

	for _, k := range keys {
		if len(keys) > 1 {
			fmt.Printf("\n---%s---\n", k)
		}
		processKey(k, gapiServices, blacklisted, falsePos)
	}

	utils.KeyLogFile.Close()
}
