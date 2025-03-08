package japan

type VersionCheckResponse struct {
	ConnectionGroups []ConnectionGroup `json:"ConnectionGroups"`
}

type ConnectionGroup struct {
	APIURL                     string                    `json:"ApiUrl"`
	BundleVersion              string                    `json:"BundleVersion"`
	CustomerServiceURL         string                    `json:"CustomerServiceUrl"`
	GatewayURL                 string                    `json:"GatewayUrl"`
	IsProductionAddressables   bool                      `json:"IsProductionAddressables"`
	KibanaLogURL               string                    `json:"KibanaLogUrl"`
	ManagementDataURL          string                    `json:"ManagementDataUrl"`
	Name                       string                    `json:"Name"`
	OverrideConnectionGroups   []OverrideConnectionGroup `json:"OverrideConnectionGroups"`
	ProhibitedWordBlackListURI string                    `json:"ProhibitedWordBlackListUri"`
	ProhibitedWordWhiteListURI string                    `json:"ProhibitedWordWhiteListUri"`
}

type OverrideConnectionGroup struct {
	AddressablesCatalogURLRoot string `json:"AddressablesCatalogUrlRoot"`
	Name                       string `json:"Name"`
}

type ServerInfoData struct {
	SkipTutorial           string `json:"SkipTutorial"`
	ServerInfoDataURL      string `json:"ServerInfoDataUrl"`
	Language               string `json:"Language"`
	DefaultConnectionGroup string `json:"DefaultConnectionGroup"`
}
