package fireblazer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetBrandIdentity(projectNumber string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://clientauthconfig.googleapis.com/v1/brands/lookupkey/brand/%s?readMask=*&%%24outputDefaults=true", projectNumber)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Origin", "https://console.cloud.google.com")
	req.Header.Set("X-Goog-Api-Key", "AIzaSyCI-zsRP85UVOi0DjtiCwWBwQ1djDy741g")

	resp, err := GetClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
