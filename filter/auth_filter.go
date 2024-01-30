package filter

import (
	"context"
	"github.com/google/wire"
	"github.com/tyeryan/l-common-util/cache"
	"lake-go/grpcclient"

	pb "github.com/tyeryan/l-protocol/go/lauth"
)

var (
	AuthFilterWireSet = wire.NewSet(
		grpcclient.ProvideLAuthClient,
		grpcclient.ProvideLAuthConfig,
		ProvideAuthFilter,
	)
)

type AuthFilter struct {
	client      pb.LAuthClient
	cacheClient cache.DistributedCache
}

func ProvideAuthFilter(ctx context.Context,
	client pb.LAuthClient,
	cacheClient cache.DistributedCache) (*AuthFilter, error) {
	return &AuthFilter{
		client:      client,
		cacheClient: cacheClient,
	}, nil
}
