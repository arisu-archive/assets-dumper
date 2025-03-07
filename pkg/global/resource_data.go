package global

import "time"

type ResourceData struct {
	ID                  int64             `json:"id"`
	MarketGameID        string            `json:"market_game_id"`
	BuildID             []int64           `json:"build_id"`
	PatchVersion        int64             `json:"patch_version"`
	Name                string            `json:"name"`
	PatchState          string            `json:"patch_state"`
	SecurityChecked     bool              `json:"security_checked"`
	MultiLanguage       bool              `json:"multi_language"`
	MultiTextureEncode  bool              `json:"multi_texture_encode"`
	MultiTextureQuality bool              `json:"multi_texture_quality"`
	Description         string            `json:"description"`
	Register            string            `json:"register"`
	RegisterDate        time.Time         `json:"register_date"`
	Updater             string            `json:"updater"`
	UpdateDate          time.Time         `json:"update_date"`
	Compress            bool              `json:"compress"`
	Size                int64             `json:"size"`
	Count               int64             `json:"count"`
	UseMultiResource    bool              `json:"use_multi_resource"`
	Category            Category          `json:"category"`
	CategoryMapping     []CategoryMapping `json:"category_mapping"`
	Resources           []Resource        `json:"resources"`
}

type Category struct {
	Lang                any      `json:"lang"`
	TextureEncodeType   any      `json:"texture_encode_type"`
	TextureQualityLevel any      `json:"texture_quality_level"`
	Group               []string `json:"group"`
}

type CategoryMapping struct {
	Group string   `json:"group"`
	Paths []string `json:"paths"`
}

type Resource struct {
	Group        string `json:"group"`
	ResourcePath string `json:"resource_path"`
	ResourceSize int64  `json:"resource_size"`
	ResourceHash string `json:"resource_hash"`
}
