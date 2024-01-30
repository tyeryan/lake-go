package apm

import (
	"context"

	"github.com/google/wire"

	"github.com/tyeryan/l-common-util/config"
	logutil "github.com/tyeryan/l-protocol/log"
)

var (
	WireSet = wire.NewSet(ProvideApmConfig)
	log     = logutil.GetLogger("apm")
)

type ApmConfig struct {
	Enable bool `configstruct:"APM_ENABLE" configdefault:"false"`
}

func ProvideApmConfig(ctx context.Context, configStore config.ConfigStore) (*ApmConfig, error) {
	apmConfig := &ApmConfig{}
	if err := configStore.GetConfig(apmConfig); err != nil {
		return nil, err
	}

	log.Debugw(ctx, "apm configured",
		"enable", apmConfig.Enable)

	return apmConfig, nil
}
