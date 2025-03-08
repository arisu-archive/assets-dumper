package resourceapi

type CatalogType string

const (
	CatalogTypeTableBundle        CatalogType = "TableBundles"
	CatalogTypeMediaResources     CatalogType = "MediaResources"
	CatalogTypeBundleDownloadInfo CatalogType = "BundleDownloadInfo"
)
