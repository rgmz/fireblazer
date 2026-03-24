package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
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

	if *key == "" {
		*key = flag.Arg(0)
		if *key == "" {
			log.Fatal("You must provide an API key. You can pass it as a named flag or as a positional flag. Usage samples: \n - \"fireblaze AIza-key\" \n - \"fireblaze --key=AIza-key\". \nTerminating.")
		}
	}
	if *dangerouslySkipVerification {
		if isInteractive || *outputFormat == "text" {
			log.Println("Skipping API key verification.")
		}
	} else if valid, err := utils.TestKeyValidity(*key); !valid {
		if err != nil {
			log.Fatalf("Error testing API key validity: %v\n. Ensure that you can connect to https://generativelanguage.googleapis.com as it's used for checking key validity. To skip primary validation (at risk of invalid results), use the --dangerouslySkipVerification flag.", err)
		}

		log.Println("Invalid API key.")
		log.Println("If you're sure the key is valid, use the --dangerouslySkipVerification flag [fireblazer --dangerouslySkipVerification AIza-KeYHere]")
		log.Printf("And submit an issue at https://github.com/bedros-p/fireblazer - include this error message:\n%v\n----", err)
		os.Exit(-1)
	} else {
		if isInteractive || *outputFormat == "text" {
			log.Println("Valid API key, proceeding.")
		}
	}

	if isInteractive || *outputFormat == "text" {
		log.Printf("Successfully loaded %d discovery endpoints from hardcoded list.", len(gapiServices))
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

	foundServices, failCount, maxTime := utils.ScanServices(*key, *referrer, gapiServices, blacklisted, falsePos, *workerCount, *timingEnabled, updateCh)

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

	utils.KeyLogFile.Close()

	log.Printf("All discovery endpoint tests completed with %d failures.", failCount)

	if *timingEnabled {
		log.Printf("Longest running service - %v\n\n\n", maxTime)
	}
}
