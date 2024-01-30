package auth

import (
	"context"
	"github.com/google/wire"
	pb "github.com/tyeryan/l-protocol/go/lauth"
)

var (
	WireSet = wire.NewSet(
		ProvideAuthHandler,
	)
)

type AuthHandler struct {
	client pb.LAuthClient
}

func ProvideAuthHandler(ctx context.Context, client pb.LAuthClient) (*AuthHandler, error) {
	return &AuthHandler{
		client: client,
	}, nil
}
