package fireblazer

import (
	_ "embed"
	"encoding/json"
)

// I named this embed.go but didnt really think it would be so central
// maybe I'd have a better per-json initializer and actually do something with metadata.go? i only kept it for that reason LMAO, i'm stuck deciding what to do and which one is better for DX
// TODO: decide where it should all go and if embed.go should just be centralized init (maybe with a better name like load_defs or something)

//go:embed data/apis.json
var rawApisJSON []byte

//go:embed data/sa_names.json
var rawSANames []byte

//go:embed data/metadata.json
var rawMetadata []byte

type APIsConfig struct {
	Active []string `json:"active"`
	Dep404 []string `json:"dep404"`
	FP     []string `json:"fp"`
}

var GoogleApiList []string

// I USED to deal with these in the outputs but in hindsight it makes no sense? Why send the request at all, we can save network res on it
// As a result I don't have to deal with allat now - if you'd like to experiment with them, feel free with these, they're still in the json for tracking

// FalsePositives []string
// var DeprecatedApis []string

var p4saProducts []string
var SANames map[string]string

func init() {
	var apiConfig APIsConfig
	_ = json.Unmarshal(rawApisJSON, &apiConfig) // data/apis.json

	GoogleApiList = apiConfig.Active
	// DeprecatedApis = apiConfig.Dep404
	// FalsePositives = apiConfig.FP

	SANames = make(map[string]string)

	_ = json.Unmarshal(rawSANames, &SANames)
	for product := range SANames {
		p4saProducts = append(p4saProducts, product)
	}

	ApiMetadata = make(map[string]ServiceMeta) // data/metadata.json
	_ = json.Unmarshal(rawMetadata, &ApiMetadata)
}
