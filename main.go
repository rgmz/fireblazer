package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	utils "github.com/bedros-p/fireblazer/utils"

	"github.com/yarlson/pin"

	"gopkg.in/yaml.v3"
)

// TODO: IAM stuff
// TODO: Print out stuff on the Pin thing and quickly wipe it like "If you can see this, your terminal size is too small or you're viewing it on a terminal that doesn't like interactive modes. Use -outputFormat=text for a better experience :)"
// TODO: Optimize further
// TODO: See if i can pre-process the JSON and bake the struct into the binary. Seems excess, but might work out
// TODO: Firebase full enumeration
// TODO: Blaze:
//	 		blaze output messages
//				flag to disable blaze entirely for individual scans
//				clarify blaze usage better and more transparently
//				clarification for dangerouslySkipVerification - this stops the -blaze from working !!

var key = flag.String("apiKey", "", "API key to scan. Can also be your first positional arg.")
var referrer = flag.String("referrer", "", "Set the referrer (Referer:) header for when an API key is restricted to a website.")
var dangerouslySkipVerification = flag.Bool("dangerouslySkipVerification", false, "Skip API key verification")
var workerCount = flag.Int("workerCount", 170, "Set the amount of worker threads to spawn for executing the requests")
var targetApi = flag.String("targetApi", "", "A single API discovery endpoint to test against. Bypasses the full scan.")

// interactive|text|json|yaml
var outputFormat = flag.String("outputFormat", "interactive", "Output format (interactive|text|json|yaml)")
var outputDetails = flag.String("outputDetails", "name", "Comma delimited list of what to include in the details (description|title|name).")
var timingEnabled = flag.Bool("findSlowService", false, "[DEBUG] Find which service took the longest to test + elapsed time. Use to file an issue for program hangs.")
var blaze = flag.Bool("blaze", false, "Enable additional aggressive recon checks (e.g., Brand Identity)")
var isInteractive = false

var scanPin = pin.New("Initializing...")
var cancel context.CancelFunc

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
	Brand         map[string]interface{}
	P4SAServices  []string
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
		if len(parts) >= 4 { // wait why am i handling this when im using splitn specifically to not deal with it i can fix this
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

func processKey(target utils.TargetKey, gapiServices []utils.Service, updateCh chan utils.ScanUpdate, logCh chan string, useGet bool) KeyResult {
	res := KeyResult{Key: target.Raw}
	if *dangerouslySkipVerification {
		if isInteractive && logCh != nil {
			logCh <- fmt.Sprintf("[%s] Skipping API key verification.", target.Raw)
		}
		res.Valid = true
	} else if valid, projectDetails, err := utils.TestKeyValidity(target.Key); !valid {
		res.Valid = false
		res.InvalidReason = err
		if updateCh != nil {
			updateCh <- utils.ScanUpdate{Key: target.Raw, CurrentRem: 0, CurrentFound: 0, ItemCleanName: "[INVALID]"}
		}
		return res
	} else {
		res.ProjectId = projectDetails.ProjectId
		if isInteractive && logCh != nil {
			logCh <- fmt.Sprintf("[%s] Valid API key, proceeding.", target.Raw)
		} else if *outputFormat == "text" {
			log.Printf("[%s] is a valid API key.", target.Raw)
		}
		res.Valid = true
	}

	foundServices, failCount, maxTime := utils.ScanServices(target, gapiServices, *workerCount, *timingEnabled, updateCh, useGet)
	res.FoundServices = foundServices
	res.FailCount = failCount
	res.MaxTime = maxTime
	return res
}

// not too sure how to handle this in the schema without bloating it up, but here's what i think
// might just have some utility `fireblazer describe`. I lowk want to make a sep tool for quick single service surface mapping, it might work better there.
type ServiceDetail struct {
	Name        string `json:"name" yaml:"name"`
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type StructuredOutput struct {
	Key                    string                 `json:"key" yaml:"key"`
	Valid                  bool                   `json:"valid" yaml:"valid"`
	InvalidReason          string                 `json:"invalid_reason,omitempty" yaml:"invalid_reason,omitempty"`
	ProjectId              string                 `json:"project_id,omitempty" yaml:"project_id,omitempty"`
	Brand                  map[string]interface{} `json:"brand,omitempty" yaml:"brand,omitempty"`
	Services               []string               `json:"services" yaml:"services"`
	ServiceDetails         []ServiceDetail        `json:"service_details,omitempty" yaml:"service_details,omitempty"`
	P4SAServices           []string               `json:"inferred_services,omitempty" yaml:"inferred_services,omitempty"`
	InferredServiceDetails []ServiceDetail        `json:"inferred_service_details,omitempty" yaml:"inferred_service_details,omitempty"`
	FailCount              int                    `json:"fail_count,omitempty" yaml:"fail_count,omitempty"`
}

func main() {
	flag.Parse()

	detailsMode := *outputDetails
	showTitle := false
	showDesc := false

	for _, p := range strings.Split(detailsMode, ",") {
		p = strings.TrimSpace(p)
		if p == "full" {
			showTitle = true
			showDesc = true
		} else if p == "title" {
			showTitle = true
		} else if p == "description" {
			showDesc = true
		}
	}

	isInteractive = *outputFormat == "interactive" || *outputFormat == ""
	if !*timingEnabled {
		log.SetFlags(0)
	}
	// utils.MultipartAllDiscoveries(*key, []string{"generativelanguage.googleapis.com", "discovery.googleapis.com"})
	// return

	if isInteractive {
		cancel = scanPin.Start(context.Background())
	}
	gapiServices := make([]utils.Service, 0)

	if *targetApi != "" {
		hostname := strings.Split(*targetApi, "/")[0]
		cleanName := strings.Split(hostname, ".")[0]

		if cleanName == "www" { // yuck 1
			parts := strings.Split(*targetApi, "/")
			if len(parts) >= 5 && parts[1] == "discovery" {
				cleanName = parts[4]
			}
		}

		discoveryUrl := *targetApi
		if !strings.HasPrefix(discoveryUrl, "http") {
			discoveryUrl = "https://" + discoveryUrl
		}

		gapiServices = append(gapiServices, utils.Service{
			CleanName:    cleanName,
			DiscoveryUrl: discoveryUrl,
		})
	} else {
		for _, raw := range utils.GoogleApiList {
			hostname := strings.Split(raw, "/")[0]
			cleanName := strings.Split(hostname, ".")[0]

			if cleanName == "www" { // yuck 1 - i dont need to handle them differently now that i decoupled the discovery rest format from the code, left this in without realizing i can kill it too
				parts := strings.Split(raw, "/")
				if len(parts) >= 5 && parts[1] == "discovery" {
					cleanName = parts[4]
				}
			}

			discoveryUrl := "https://" + raw

			gapiServices = append(gapiServices, utils.Service{
				CleanName:    cleanName,
				DiscoveryUrl: discoveryUrl,
			})
		}
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
		if *targetApi != "" {
			log.Printf("Using single target API: %s", *targetApi)
		} else {
			log.Printf("Successfully loaded %d discovery endpoints from built-in program list.", len(gapiServices))
		}
	}

	var updateCh chan utils.ScanUpdate
	var logCh chan string
	var updateDone chan struct{}

	if isInteractive { // I feel this interactive display thing kinda deserves its own function because holy indents
		defer cancel()

		updateCh = make(chan utils.ScanUpdate, *workerCount*len(keys))
		logCh = make(chan string, len(keys)*3)
		updateDone = make(chan struct{})

		go func() {
			totalRemMap := make(map[string]int)
			totalFoundMap := make(map[string]int)

			for _, k := range keys {
				tk := parseTargetKey(k, *referrer)
				totalRemMap[tk.Raw] = len(gapiServices)
			}

			for updateCh != nil || logCh != nil {
				select {
				case update, ok := <-updateCh:
					if !ok {
						updateCh = nil
						continue
					}
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

					// i feel i should handle more of the interactive segment here. like at the very least, the "remaining" section should be managed in main.go, i feel anyone looking at this would be confused

					scanPin.UpdateMessage(fmt.Sprintf("Keys %d | Found %d | Rem %d | Scanning %s", len(keys), totalFound, totalRem, update.ItemCleanName))
				case msg, ok := <-logCh:
					if !ok {
						logCh = nil
						continue
					}
					fmt.Printf("\x1b[2K\r%s\n", msg)
				}
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
			res := processKey(target, gapiServices, updateCh, logCh, *targetApi != "")

			if res.Valid && res.ProjectId != "" && (*blaze || (isInteractive && len(keys) == 1)) {
				brand, err := utils.GetBrandIdentity(res.ProjectId)
				if err == nil && brand != nil {
					res.Brand = brand
				}

				saServices, err := utils.EnumerateServiceAccounts(res.ProjectId, *workerCount)
				if err == nil && len(saServices) > 0 {
					res.P4SAServices = saServices
				}
			}
			results[i] = res
		}(i, k)
	}

	wg.Wait()

	if isInteractive {
		if updateCh != nil {
			close(updateCh)
		}
		if logCh != nil {
			close(logCh)
		}
		<-updateDone
		scanPin.Stop("Scan complete!")
	} else if *outputFormat == "text" {
		log.Println("Scan complete!")
	}

	if *outputFormat == "json" || *outputFormat == "yaml" {
		// I really need to see if i could clean up my output format logic :/

		var structuredResults []StructuredOutput

		for _, res := range results {
			out := StructuredOutput{
				Key:          res.Key,
				Valid:        res.Valid,
				ProjectId:    res.ProjectId,
				Brand:        res.Brand,
				FailCount:    res.FailCount,
				Services:     []string{},
				P4SAServices: res.P4SAServices,
			}

			if !res.Valid && res.InvalidReason != nil {
				out.InvalidReason = res.InvalidReason.Error()
			}

			var sDetails []ServiceDetail

			for _, service := range res.FoundServices {
				serviceName := service + ".googleapis.com"
				out.Services = append(out.Services, serviceName)

				if showTitle || showDesc {
					meta, hasMeta := utils.ApiMetadata[service]
					detail := ServiceDetail{Name: serviceName}

					if hasMeta {
						if showTitle {
							detail.Title = meta.Title
						}
						if showDesc {
							detail.Description = meta.Summary // insane level of nesting I think something's wrong here, or i can use more guard clauses thruu my code
						}
					}

					sDetails = append(sDetails, detail)
				}
			}

			if len(sDetails) > 0 {
				out.ServiceDetails = sDetails
			}

			var inferredDetails []ServiceDetail
			for _, saSvc := range res.P4SAServices {
				detail := ServiceDetail{Name: saSvc}
				if showTitle || showDesc {
					if name, exists := utils.SANames[saSvc]; exists {
						detail.Title = name
					}
				}
				inferredDetails = append(inferredDetails, detail)
			}

			if len(inferredDetails) > 0 {
				out.InferredServiceDetails = inferredDetails
			}

			structuredResults = append(structuredResults, out)
		}

		if *outputFormat == "json" {
			jsonData, err := json.Marshal(structuredResults)

			if err != nil {
				log.Fatalf("Error marshaling JSON: %v", err)
			}

			fmt.Println(string(jsonData))
		} else {
			yamlData, err := yaml.Marshal(structuredResults)

			if err != nil {
				log.Fatalf("Error marshaling YAML: %v", err) // i do wonder if i can maintain a more clean app state and the final result output can be a handler for the final output. would be much cleaner, i already put in so much work for a good app state for this to work
			}
			// wait actually why dont i move it? the app state seems to have everything to begin with, i think it was just misguided optimization cope. there are better things to focus on for opt than this. DX matters when we have this insanely monolithic main go
			fmt.Println(string(yamlData))
		}
	} else {

		if *targetApi != "" {
			fmt.Printf("\nTARGET: %s\n", *targetApi)
		}

		for _, res := range results {
			if *targetApi != "" {
				status := "❌"
				if len(res.FoundServices) > 0 {
					status = "✅"
				}
				fmt.Printf("%s : %s\n", res.Key, status)
				continue
			}

			if len(keys) > 1 {
				fmt.Printf("\n---%s---\n", res.Key)
			}
			if !res.Valid {
				log.Printf("Invalid API key: %s\nError testing validity: %v\nIf you're sure the key is valid, use the --dangerouslySkipVerification flag.", res.Key, res.InvalidReason)
				continue
			}

			if res.Brand != nil {
				log.Printf("\nOAuth Client Screen Details")
				if displayName, ok := res.Brand["displayName"]; ok && displayName != "" {
					log.Printf(" - App Name: %v", displayName)
				}
				if supportEmail, ok := res.Brand["supportEmail"]; ok && supportEmail != "" {
					log.Printf(" - Project Admin / Support Email: %v", supportEmail)
				}
				if homePageUrl, ok := res.Brand["homePageUrl"]; ok && homePageUrl != "" {
					log.Printf(" - Homepage: %v", homePageUrl)
				}
			}

			log.Printf("\nAPIs available to this API key with project ID %s:", res.ProjectId)

			for _, service := range res.FoundServices {

				baseMsg := fmt.Sprintf(" - %s.googleapis.com", service)
				meta, hasMeta := utils.ApiMetadata[service]

				details := ""
				if hasMeta {
					if showTitle && meta.Title != "" { // this made perfect sense when i wrote it but i dont like it because my code is littered with "" and " - "
						details += meta.Title
					}
					if showDesc && meta.Summary != "" {
						if details != "" {
							details += " - "
						}
						details += meta.Summary
					}
				}
				if details != "" {
					log.Printf("%s\n   ^-- %s", baseMsg, details)
				} else {
					log.Printf(baseMsg)
				}
			}

			if len(res.P4SAServices) > 0 {
				log.Printf("\n[Additional Recon] Inferred Services via Service Accounts:")
				for _, saSvc := range res.P4SAServices {
					if name, exists := utils.SANames[saSvc]; exists {
						log.Printf(" - %s (%s)", saSvc, name)
					} else {
						log.Printf(" - %s", saSvc)
					}
				}
				log.Printf("")
			}

			log.Printf("All discovery endpoint tests completed with %d failures.", res.FailCount)

			if *timingEnabled {
				log.Printf("Longest running service - %v\n\n\n", res.MaxTime)
			}
		}
	}

	utils.KeyLogFile.Close()
}
