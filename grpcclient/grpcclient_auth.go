package grpcclient

import (
	"context"
	config "github.com/tyeryan/l-common-util/config"
	grpcclient "github.com/tyeryan/l-common-util/grpcclient"
	"github.com/tyeryan/l-protocol/go/lauth"
	. "github.com/tyeryan/l-protocol/log"
	"time"
)

type LAuthConfig struct {
	LAuthAddr              string `configstruct:"GRPC_CLIENT_CONFIG_L_AUTH"`
	LAuthTimeoutInSec      int32  `configdefault:"30" configstruct:"GRPC_CLIENT_CONFIG_L_AUTH_TIMEOUT_IN_SEC,omitempty"`
	LAuthRetryBackoffInSec int32  `configdefault:"1" configstruct:"GRPC_CLIENT_CONFIG_L_AUTH_RETRY_BACKOFF_IN_SEC,omitempty"`
}

func ProvideLAuthConfig(ctx context.Context, configStore config.ConfigStore) (*LAuthConfig, error) {
	cnf := LAuthConfig{}
	if err := configStore.GetConfig(&cnf); err != nil {
		return nil, err
	}
	return &cnf, nil
}

func ProvideLAuthClient(ctx context.Context, cnf *LAuthConfig) (lauth.LAuthClient, error) {
	log := GetLogger("ProvideCAuthClient")
	conn, err := grpcclient.NewGRPCConnection(cnf.LAuthAddr, grpcclient.WithTimeout(time.Duration(cnf.LAuthTimeoutInSec)), grpcclient.WithRetryBackoff(time.Duration(cnf.LAuthRetryBackoffInSec)))
	if err != nil {
		log.Errore(ctx, "connect to l-auth service found error", err)
		return nil, err
	}
	return lauth.NewLAuthClient(conn), nil
}
