package resources

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/arisu-archive/assets-dumper/pkg/global"
	"github.com/arisu-archive/assets-dumper/pkg/japan"
)

type Client interface {
	GetResource(ctx context.Context, resourcePath string) ([]byte, error)
	GetResources(ctx context.Context, filter string) ([]string, error)
	GetVersion(ctx context.Context) (string, error)
}

func NewClient(server Server) (Client, error) {
	client := resty.New()
	switch server {
	case ServerGlobal:
		return global.NewClient(client), nil
	case ServerJapan:
		return japan.NewClient(client), nil
	default:
		return nil, fmt.Errorf("unknown server: %s", server)
	}
}
