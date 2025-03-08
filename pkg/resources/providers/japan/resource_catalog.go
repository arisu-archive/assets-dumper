package japan

type MediaType int32

const (
	MediaTypeNone MediaType = iota
	MediaTypeAudio
	MediaTypeVideo
	MediaTypeTexture
)

type MediaCatalog struct {
	MediaBundles map[string]MediaBundle `json:"MediaBundles"`
}

type MediaBundle struct {
	Path            string    `json:"Path"`
	FileName        string    `json:"FileName"`
	Bytes           int64     `json:"Bytes"`
	Crc             int64     `json:"Crc"`
	IsPrologue      bool      `json:"IsPrologue"`
	IsSplitDownload bool      `json:"IsSplitDownload"`
	MediaType       MediaType `json:"MediaType"`
}

type TableCatalog struct {
	TableBundles map[string]TableBundle `json:"TableBundles"`
}

type TableBundle struct {
	Path            string   `json:"Path"`
	Bytes           int64    `json:"Bytes"`
	Crc             int64    `json:"Crc"`
	IsInBuild       bool     `json:"IsInBuild"`
	IsChanged       bool     `json:"IsChanged"`
	IsPrologue      bool     `json:"IsPrologue"`
	IsSplitDownload bool     `json:"IsSplitDownload"`
	Includes        []string `json:"Includes"`
}

type BundleDownloadInfo struct {
	Files []BundleFile `json:"BundleFiles"`
}

type BundleFile struct {
	Name            string `json:"Name"`
	Size            int64  `json:"Size"`
	IsPrologue      bool   `json:"IsPrologue"`
	Crc             int64  `json:"Crc"`
	IsSplitDownload bool   `json:"IsSplitDownload"`
}
