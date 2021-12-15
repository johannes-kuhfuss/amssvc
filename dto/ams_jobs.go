package dto

type AmsJobReq struct {
	Properties Properties `json:"properties"`
}
type Input struct {
	OdataType string   `json:"@odata.type"`
	BaseURI   string   `json:"baseUri"`
	Files     []string `json:"files"`
}
type Outputs struct {
	OdataType string `json:"@odata.type"`
	AssetName string `json:"assetName"`
}
type Properties struct {
	Input    Input     `json:"input"`
	Outputs  []Outputs `json:"outputs"`
	Priority string    `json:"priority"`
}
