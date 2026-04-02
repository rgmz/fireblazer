package fireblazer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type EmailTemplate struct {
	From            string `json:"from,omitempty"`
	FromDisplayName string `json:"fromDisplayName,omitempty"`
	Subject         string `json:"subject,omitempty"`
	Format          string `json:"format,omitempty"`
	Body            string `json:"body,omitempty"`
	ReplyTo         string `json:"replyTo,omitempty"`
}

type IdpConfig struct {
	Enabled              bool     `json:"enabled,omitempty"`
	WhitelistedAudiences []string `json:"whitelistedAudiences,omitempty"`
	Secret               string   `json:"secret,omitempty"`
	ExperimentPercent    int32    `json:"experimentPercent,omitempty"`
	ClientId             string   `json:"clientId,omitempty"`
	Provider             string   `json:"provider,omitempty"`
}

type ProjectDetails struct {
	ProjectId                   string         `json:"projectId,omitempty"`
	AuthorizedDomains           []string       `json:"authorizedDomains,omitempty"`
	UseEmailSending             bool           `json:"useEmailSending,omitempty"`
	EnableAnonymousUser         bool           `json:"enableAnonymousUser,omitempty"`
	AllowPasswordUser           bool           `json:"allowPasswordUser,omitempty"`
	ApiKey                      string         `json:"apiKey,omitempty"`
	DynamicLinksDomain          string         `json:"dynamicLinksDomain,omitempty"`
	ChangeEmailTemplate         *EmailTemplate `json:"changeEmailTemplate,omitempty"`
	ResetPasswordTemplate       *EmailTemplate `json:"resetPasswordTemplate,omitempty"`
	LegacyResetPasswordTemplate *EmailTemplate `json:"legacyResetPasswordTemplate,omitempty"`
	VerifyEmailTemplate         *EmailTemplate `json:"verifyEmailTemplate,omitempty"`
	IdpConfig                   []IdpConfig    `json:"idpConfig,omitempty"`
}

type IdentityToolkitResponse struct {
	ProjectDetails
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Details []struct {
			Metadata struct {
				Consumer string `json:"consumer"`
			} `json:"metadata"`
		} `json:"details"`
	} `json:"error"`
}

const keyCheckEndpoint = "https://www.googleapis.com/identitytoolkit/v3/relyingparty/getProjectConfig"

// Contains all general google-specific shenanigans that don't belong elsewhere (behavior that has lore to it kinda)

func TestKeyValidity(apiKey string) (bool, *ProjectDetails, error) {
	sharedClient := GetClient()
	req, _ := http.NewRequest("GET", AppendAPIKeyToURL(keyCheckEndpoint, apiKey), nil)
	resp, err := ReqWithBackoff(req, sharedClient)

	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return false, nil, err
	}

	defer resp.Body.Close()

	var result IdentityToolkitResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Error != nil {
		errorMarshal, err := json.Marshal(result)
		errorString := string(errorMarshal) + " ---- HTTP STATUS " + resp.Status
		if result.Error.Code == 400 || result.Error.Message == "API key not valid. Please pass a valid API key." {
			if err != nil {
				return false, nil, fmt.Errorf("DOUBLE WHAMMY : API Key not valid, JSON marshal for error message failed too. Error->Message: %s", result.Error.Message)
			}
			return false, nil, fmt.Errorf("%v", errorString)
		} else if resp.StatusCode == 403 {
			// BY THE WAY!!! this doesnt mean the project doesn't have access to the API. this method could just be blocked

			details := &ProjectDetails{}
			for _, d := range result.Error.Details {
				if d.Metadata.Consumer != "" {
					details.ProjectId = strings.TrimPrefix(d.Metadata.Consumer, "projects/")
					break
				}
			}

			return true, details, nil
		}
		return false, nil, fmt.Errorf("Unknown error checking key validity: %v", errorString)
	}

	return true, &result.ProjectDetails, nil
}

// Parse query params and append the API key to it.
func AppendAPIKeyToURL(apiUrl string, apiKey string) string {
	httpURL, _ := url.Parse(apiUrl)
	values := httpURL.Query()
	values.Add("key", apiKey)
	return httpURL.Scheme + "://" + httpURL.Host + httpURL.Path + "?" + values.Encode()
}
