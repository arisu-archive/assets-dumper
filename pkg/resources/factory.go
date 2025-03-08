package resources

import (
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
	"github.com/arisu-archive/assets-dumper/pkg/resources/providers/global"
	"github.com/arisu-archive/assets-dumper/pkg/resources/providers/japan"
)

func NewClient(server resourceapi.Server) (resourceapi.Client, error) {
	client := resty.New()
	switch server {
	case resourceapi.ServerGlobal:
		return global.NewClient(client), nil
	case resourceapi.ServerJapan:
		return japan.NewClient(client), nil
	default:
		return nil, fmt.Errorf("unknown server: %s", server)
	}
}
