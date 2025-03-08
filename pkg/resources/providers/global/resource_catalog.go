package global

type MediaType int32

const (
	MediaTypeNone MediaType = iota
	MediaTypeAudio
	MediaTypeVideo
	MediaTypeTexture
)

type StorageType int32

const (
	StorageTypeNone StorageType = iota
	StorageTypeInBuild
	StorageTypePreload
	StorageTypeGameData
)

type MediaCatalog struct {
	MediaBundles map[string]MediaBundle `json:"MediaBundles"`
	Catalog      map[string]MediaBundle `json:"Catalog"`
}

type MediaBundle struct {
	Path        string      `json:"Path"`
	StorageType StorageType `json:"StorageType"`
	MediaType   MediaType   `json:"MediaType"`
}

type TableCatalog struct {
	TableBundles map[string]TableBundle `json:"TableBundles"`
	Catalog      map[string]TableBundle `json:"Catalog"`
}

type TableBundle struct {
	Path       string   `json:"Path"`
	IsPrologue bool     `json:"IsPrologue"`
	Includes   []string `json:"Includes"`
}
