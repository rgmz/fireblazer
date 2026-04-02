package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"

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

var scanPin = pin.New("Initializing...")
var cancel = scanPin.Start(context.Background())

type APIDetails struct {
	Description string
	Title       string
}

type KeyResult struct {
	Key           string
	ProjectId     string
	Valid         bool
	InvalidReason error
	FoundServices []string
	FailCount     int
	MaxTime       *utils.ElapsedCombo
}

func parseTargetKey(raw string, globalRef string) utils.TargetKey {
	tk := utils.TargetKey{
		Raw:      raw,
		Key:      raw,
		Referrer: globalRef,
	}

	if strings.HasPrefix(raw, "xios:") {
		parts := strings.SplitN(raw, ":", 3)
		if len(parts) >= 3 {
			tk.Key = parts[1]
			tk.IosBundleId = parts[2]
			tk.Referrer = ""
		}
	} else if strings.HasPrefix(raw, "xandroid:") {
		parts := strings.SplitN(raw, ":", 4)
		if len(parts) >= 4 {
			tk.Key = parts[1]
			tk.AndroidPkg = parts[2]
			tk.AndroidCert = parts[3]
			tk.Referrer = ""
		}
	} else if strings.HasPrefix(raw, "xref:") {
		parts := strings.SplitN(raw, ":", 3)
		if len(parts) >= 3 {
			tk.Key = parts[1]
			tk.Referrer = parts[2]
		}
	}
	return tk
}

func processKey(target utils.TargetKey, gapiServices []utils.Service, blacklisted []string, falsePos []string, updateCh chan utils.ScanUpdate) KeyResult {
	res := KeyResult{Key: target.Raw}
	if *dangerouslySkipVerification {
		if isInteractive {
			scanPin.UpdateMessage(fmt.Sprintf("[%s] Skipping API key verification.", target.Raw))
		}
		res.Valid = true
	} else if valid, projectDetails, err := utils.TestKeyValidity(target.Key); !valid {
		res.Valid = false
		res.InvalidReason = err
		return res
	} else {
		res.ProjectId = projectDetails.ProjectId
		if isInteractive {
			// there's probably a better way to make a separate display, but regular logs overlap on the same line.
			scanPin.Stop(fmt.Sprintf("[%s] Valid API key, proceeding.", target.Raw))
			scanPin.Start(context.Background())
		} else if *outputFormat == "text" {
			log.Printf("[%s] is a valid API key.", target.Raw)
		}
		res.Valid = true
	}

	foundServices, failCount, maxTime := utils.ScanServices(target, gapiServices, blacklisted, falsePos, *workerCount, *timingEnabled, updateCh)
	res.FoundServices = foundServices
	res.FailCount = failCount
	res.MaxTime = maxTime
	return res
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
		var discoveryUrl string
		if cleanName == "drive" { // HOTFIX! WILL INVESTIGATE PROPERLY RN!
			discoveryUrl = "https://www.googleapis.com/discovery/v1/apis/drive/v3/rest"
		} else {
			discoveryUrl = "https://" + hostname + "/$discovery/rest"

		}

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
		log.Fatal("You must provide at least one API key. You can pass it as a named flag or as positional arguments. Usage samples: \n - \"fireblazer AIza-key1 AIza-key2\" \n - \"fireblazer --apiKey=AIza-key\". \nTerminating.")
	}

	if isInteractive || *outputFormat == "text" {
		log.Printf("Successfully loaded %d discovery endpoints from hardcoded list.", len(gapiServices))
	}

	var updateCh chan utils.ScanUpdate
	var updateDone chan struct{}

	if isInteractive {
		defer cancel()

		updateCh = make(chan utils.ScanUpdate, *workerCount*len(keys))
		updateDone = make(chan struct{})

		go func() {
			totalRemMap := make(map[string]int)
			totalFoundMap := make(map[string]int)

			for update := range updateCh {
				totalRemMap[update.Key] = update.CurrentRem
				totalFoundMap[update.Key] = update.CurrentFound

				totalRem := 0
				totalFound := 0
				for _, rem := range totalRemMap {
					totalRem += rem
				}
				for _, f := range totalFoundMap {
					totalFound += f
				}

				scanPin.UpdateMessage(fmt.Sprintf("Keys %d | Found %d | Rem %d | Scanning %v", len(keys), totalFound, totalRem, update.ItemCleanName))
			}
			close(updateDone)
		}()
	}

	var wg sync.WaitGroup
	results := make([]KeyResult, len(keys))

	for i, k := range keys {
		wg.Add(1)
		go func(i int, rawKey string) {
			defer wg.Done()
			target := parseTargetKey(rawKey, *referrer)
			results[i] = processKey(target, gapiServices, blacklisted, falsePos, updateCh)
		}(i, k)
	}

	wg.Wait()

	if isInteractive {
		if updateCh != nil {
			close(updateCh)
		}
		<-updateDone
		scanPin.Stop("Scan complete!")
	} else if *outputFormat == "text" {
		log.Println("Scan complete!")
	}

	for _, res := range results {
		if len(keys) > 1 {
			fmt.Printf("\n---%s---\n", res.Key)
		}
		if !res.Valid {
			log.Printf("Invalid API key: %s\nError testing validity: %v\nIf you're sure the key is valid, use the --dangerouslySkipVerification flag.", res.Key, res.InvalidReason)
			continue
		}

		log.Printf("APIs available to this API key with project ID %s:", res.ProjectId)

		for _, service := range res.FoundServices {
			if slices.Contains(falsePos, service) {
				// log.Printf(" - %s.googleapis.com (false positive)", service)
			} else {
				log.Printf(" - %s.googleapis.com", service)
			}
		}

		log.Printf("All discovery endpoint tests completed with %d failures.", res.FailCount)

		if *timingEnabled {
			log.Printf("Longest running service - %v\n\n\n", res.MaxTime)
		}
	}

	utils.KeyLogFile.Close()
}
