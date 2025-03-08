package global

type VersionCheckResponse struct {
	APIVersion         string `json:"api_version"`
	Language           string `json:"language"`
	LatestBuildNumber  string `json:"latest_build_number"`
	LatestBuildVersion string `json:"latest_build_version"`
	MarketGameID       string `json:"market_game_id"`
	MinBuildNumber     string `json:"min_build_number"`
	MinBuildVersion    string `json:"min_build_version"`
	Patch              Patch  `json:"patch"`
}

type Patch struct {
	BdiffPath    []map[string]string `json:"bdiff_path"`
	PatchVersion int64               `json:"patch_version"`
	ResourcePath string              `json:"resource_path"`
}
