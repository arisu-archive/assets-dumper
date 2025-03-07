package japan

type MediaType int32

const (
	MediaTypeNone MediaType = iota
	MediaTypeAudio
	MediaTypeVideo
	MediaTypeTexture
)

type MediaCatalog struct {
	MediaBundles map[string]MediaBundle
}

type MediaBundle struct {
	Path            string
	FileName        string
	Bytes           int64
	Crc             int64
	IsPrologue      bool
	IsSplitDownload bool
	MediaType       MediaType
}

type TableCatalog struct {
	TableBundles map[string]TableBundle
}

type TableBundle struct {
	Path            string
	Bytes           int64
	Crc32           int64
	IsInBuild       bool
	IsChanged       bool
	IsPrologue      bool
	IsSplitDownload bool
	Includes        []string
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
