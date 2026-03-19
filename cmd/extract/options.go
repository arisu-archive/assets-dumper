package extract

import (
	"encoding/base64"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

type options struct {
	server     string
	inputPath  string
	outputPath string
	key        string
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

	// It must be base64-encoded string if the key is provided.
	if o.key != "" {
		if _, err := base64.StdEncoding.DecodeString(o.key); err != nil {
			return ErrInvalidKey
		}
	}

	return nil
}
