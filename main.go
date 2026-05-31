package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	lib "github.com/bedros-p/fireblazer/lib"

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
	MaxTime       *lib.ElapsedCombo
	Brand         map[string]interface{}
	P4SAServices  []string
}

func parseTargetKey(raw string, globalRef string) lib.TargetKey {
	tk := lib.TargetKey{
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

func processKey(target lib.TargetKey, gapiServices []lib.Service, updateCh chan lib.ScanUpdate, logCh chan string, useGet bool) KeyResult {
	res := KeyResult{Key: target.Raw}
	if *dangerouslySkipVerification {
		if isInteractive && logCh != nil {
			logCh <- fmt.Sprintf("[%s] Skipping API key verification.", target.Raw)
		}
		res.Valid = true
	} else if valid, projectDetails, err := lib.TestKeyValidity(target.Key); !valid {
		res.Valid = false
		res.InvalidReason = err
		if updateCh != nil {
			updateCh <- lib.ScanUpdate{Key: target.Raw, WasFound: false, ItemCleanName: "[INVALID]"}
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

	foundServices, failCount, maxTime := lib.ScanServices(target, gapiServices, *workerCount, *timingEnabled, updateCh, useGet)
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
	gapiServices := loadServices(*targetApi)

	rawKeys := []string{}
	if *key != "" {
		rawKeys = append(rawKeys, *key)
	}
	rawKeys = append(rawKeys, flag.Args()...)

	keys := []string{}
	seenKeys := make(map[string]bool)
	for _, k := range rawKeys {
		if !seenKeys[k] {
			seenKeys[k] = true
			keys = append(keys, k)
		}
	}

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

	var updateCh chan lib.ScanUpdate
	var logCh chan string
	var updateDone chan struct{}

	if isInteractive {
		defer cancel()
		updateCh, logCh, updateDone = startInteractiveDisplay(keys, len(gapiServices), *referrer)
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
				brand, err := lib.GetBrandIdentity(res.ProjectId)
				if err == nil && brand != nil {
					res.Brand = brand
				}

				saServices, err := lib.EnumerateServiceAccounts(res.ProjectId, *workerCount)
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

	var outputData []byte
	var err error

	if *outputFormat == "json" {
		outputData, err = json.MarshalIndent(marshalStructured(results, showTitle, showDesc), "", "  ")
	} else if *outputFormat == "yaml" {
		outputData, err = yaml.Marshal(marshalStructured(results, showTitle, showDesc))
	} else {
		outputData = marshalText(results, keys, *targetApi, *timingEnabled, showTitle, showDesc)
	}

	if err != nil {
		log.Fatalf("Error marshaling output: %v", err)
	}

	fmt.Println(string(outputData))

	lib.KeyLogFile.Close()
}

func marshalStructured(results []KeyResult, showTitle bool, showDesc bool) []StructuredOutput {
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
				meta, hasMeta := lib.ApiMetadata[service]
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
				if name, exists := lib.SANames[saSvc]; exists {
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
	return structuredResults
}

func marshalText(results []KeyResult, keys []string, targetApi string, timingEnabled bool, showTitle bool, showDesc bool) []byte {
	var buf bytes.Buffer

	if targetApi != "" {
		buf.WriteString(fmt.Sprintf("\nTARGET: %s\n", targetApi))
	}

	for _, res := range results {
		if targetApi != "" {
			status := "❌"
			if len(res.FoundServices) > 0 {
				status = "✅"
			}
			buf.WriteString(fmt.Sprintf("%s : %s\n", res.Key, status))
			continue
		}

		if len(keys) > 1 {
			buf.WriteString(fmt.Sprintf("\n---%s---\n", res.Key))
		}
		if !res.Valid {
			buf.WriteString(fmt.Sprintf("Invalid API key: %s\nError testing validity: %v\nIf you're sure the key is valid, use the --dangerouslySkipVerification flag.\n", res.Key, res.InvalidReason))
			continue
		}

		if res.Brand != nil {
			buf.WriteString("\nOAuth Client Screen Details\n")
			if displayName, ok := res.Brand["displayName"]; ok && displayName != "" {
				buf.WriteString(fmt.Sprintf(" - App Name: %v\n", displayName))
			}
			if supportEmail, ok := res.Brand["supportEmail"]; ok && supportEmail != "" {
				buf.WriteString(fmt.Sprintf(" - Project Admin / Support Email: %v\n", supportEmail))
			}
			if homePageUrl, ok := res.Brand["homePageUrl"]; ok && homePageUrl != "" {
				buf.WriteString(fmt.Sprintf(" - Homepage: %v\n", homePageUrl))
			}
		}

		buf.WriteString(fmt.Sprintf("\nAPIs available to this API key with project ID %s:\n", res.ProjectId))

		for _, service := range res.FoundServices {
			baseMsg := fmt.Sprintf(" - %s.googleapis.com", service)
			meta, hasMeta := lib.ApiMetadata[service]

			details := ""
			if hasMeta {
				if showTitle && meta.Title != "" {
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
				buf.WriteString(fmt.Sprintf("%s\n   ^-- %s\n", baseMsg, details))
			} else {
				buf.WriteString(fmt.Sprintf("%s\n", baseMsg))
			}
		}

		if len(res.P4SAServices) > 0 {
			buf.WriteString("\n[Additional Recon] Inferred Services via Service Accounts:\n")
			for _, saSvc := range res.P4SAServices {
				if name, exists := lib.SANames[saSvc]; exists {
					buf.WriteString(fmt.Sprintf(" - %s (%s)\n", saSvc, name))
				} else {
					buf.WriteString(fmt.Sprintf(" - %s\n", saSvc))
				}
			}
			buf.WriteString("\n")
		}

		buf.WriteString(fmt.Sprintf("All discovery endpoint tests completed with %d failures.\n", res.FailCount))

		if timingEnabled {
			buf.WriteString(fmt.Sprintf("Longest running service - %v\n\n\n", res.MaxTime))
		}
	}

	return buf.Bytes()
}

func loadServices(targetApi string) []lib.Service {
	var services []lib.Service

	if targetApi != "" {
		hostname := strings.Split(targetApi, "/")[0]
		cleanName := strings.Split(hostname, ".")[0]
		discoveryUrl := targetApi
		if !strings.HasPrefix(discoveryUrl, "http") {
			discoveryUrl = "https://" + discoveryUrl
		}
		services = append(services, lib.Service{
			CleanName:    cleanName,
			DiscoveryUrl: discoveryUrl,
		})
	} else {
		for _, raw := range lib.GoogleApiList {
			hostname := strings.Split(raw, "/")[0]
			cleanName := strings.Split(hostname, ".")[0]
			services = append(services, lib.Service{
				CleanName:    cleanName,
				DiscoveryUrl: "https://" + raw,
			})
		}
	}
	return services
}

func startInteractiveDisplay(keys []string, totalServices int, globalReferrer string) (chan lib.ScanUpdate, chan string, chan struct{}) {
	updateCh := make(chan lib.ScanUpdate, *workerCount*len(keys))
	logCh := make(chan string, len(keys)*3)
	updateDone := make(chan struct{})

	go func() {
		totalRem := totalServices * len(keys)
		totalFound := 0

		for updateCh != nil || logCh != nil {
			select {
			case update, ok := <-updateCh:
				if !ok {
					updateCh = nil
					continue
				}

				totalRem--
				if update.WasFound {
					totalFound++
				}

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

	return updateCh, logCh, updateDone
}
