package fireblazer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

var p4saProducts = []string{
	"aiplatform", "alloydb", "anthos", "anthosidentityservice", "anthosmulticloud",
	"apigee", "apigateway", "appengine", "artifactregistry", "backupdr",
	"baremetalsolution", "batch", "beyondcorp", "biglake", "bigquerydatatransfer",
	"bigquerydatapolicy", "bigquerymigration", "bigqueryreservation", "bigtableadmin",
	"binaryauthorization", "bms", "certificateauthority", "chromepolicy", "cloudasset",
	"cloudbuild", "clouddebugger", "clouddeploy", "cloudfunctions", "cloudiot",
	"cloudkms", "cloudprofiler", "cloudresourcemanager", "cloudscheduler", "cloudsearch",
	"cloudsupport", "cloudtasks", "cloudtrace", "composer", "compute", "connectors",
	"contactcenterinsights", "containeranalysis", "containerregistry", "datacatalog",
	"dataflow", "datafusion", "datalabeling", "datamigration", "dataplex", "dataproc",
	"datastore", "datastream", "deploymentmanager", "dialogflow", "dlp", "dns",
	"documentai", "domains", "essentialcontacts", "eventarc", "file", "firebaseappcheck",
	"firebasedatabase", "firebasehosting", "firebaseml", "firebaseremoteconfig",
	"firebasestorage", "firestore", "gameservices", "genomics", "gkeconnect", "gkehub",
	"gkemulticloud", "healthcare", "iam", "iap", "iap-corp", "identitytoolkit", "ids",
	"krmapihosting", "lifesciences", "logging", "managedidentities", "memcache",
	"metastore", "ml", "monitoring", "networkconnectivity", "networkmanagement",
	"networksecurity", "networkservices", "notebooks", "osconfig", "oslogin",
	"policytroubleshooter", "privateca", "pubsub", "recaptchaenterprise", "redis",
	"retail", "run", "secretmanager", "securitycenter", "serviceconsumermanagement",
	"servicedirectory", "servicemanagement", "serviceusage", "sourcerepo", "spanner",
	"speech", "sqladmin", "storage", "storagetransfer", "tpu", "translate", "vmmigration",
	"vmwareengine", "vpcaccess", "websecurityscanner", "webrisk", "workflows",
}

func EnumerateServiceAccounts(projectNum string, workerCount int) ([]string, error) {
	var mu sync.Mutex
	var foundProducts []string

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(workerCount)

	for _, product := range p4saProducts {
		product := product // do NOT question my sanity
		g.Go(func() error {
			email := fmt.Sprintf("service-%s@gcp-sa-%s.iam.gserviceaccount.com", projectNum, product)
			url := fmt.Sprintf("https://iam.googleapis.com/v1/projects/%s/serviceAccounts/%s", projectNum, email)

			for retries := 0; retries < 3; retries++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					return nil
				}

				// this is actually an unauthed project thing LMAO you could use this for any project ID with no key at all
				resp, err := ReqHeaderOnly(*req, "", false)
				if err != nil {
					return nil
				}

				statusCode := resp.StatusCode
				resp.Body.Close()

				if statusCode == 200 || statusCode == 403 {
					mu.Lock()
					foundProducts = append(foundProducts, product)
					mu.Unlock()
					return nil
				} else if statusCode == 502 || statusCode == 503 || statusCode == 429 {
					time.Sleep(2 * time.Second) // a 429 should NEVER happen. I mean, NEVER. It's an unauthed resource. Still, I've dealt with too many edgecases with Drive, and want to ensure nothing like that happens again.
					continue                    // never thought i'd need a retry...
				} else {
					return nil
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return foundProducts, nil
}
