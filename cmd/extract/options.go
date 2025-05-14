package extract

import "github.com/arisu-archive/assets-dumper/pkg/resourceapi"

type options struct {
	server     string
	inputPath  string
	outputPath string
}

func (o *options) Validate() error {
	if o.inputPath == "" {
		return ErrInvalidInputPath
	}

	if o.outputPath == "" {
		return ErrInvalidOutputPath
	}

	if o.server == "" || !resourceapi.GetServer(o.server).IsValid() {
		return ErrInvalidServer
	}

	return nil
}
