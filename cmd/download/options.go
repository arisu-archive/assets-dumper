package download

type options struct {
	server         string
	output         string
	filter         string
	maxConcurrency int
	version        string
	patchVersion   string
}
