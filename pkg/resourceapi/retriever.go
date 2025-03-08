package resourceapi

import "context"

type Retriever interface {
	CollectAllResources(ctx context.Context) ([]Resource, error)
}
