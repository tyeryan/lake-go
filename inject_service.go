//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"github.com/tyeryan/l-common-util/apm"
	"github.com/tyeryan/l-common-util/cache"
	"github.com/tyeryan/l-common-util/config"
	"lake-go/filter"
	"lake-go/router"
	"net/http"
)

func injectRoutes(ctx context.Context) (http.Handler, error) {
	panic(wire.Build(
		config.WireSet,
		apm.WireSet,
		cache.WireSet,
		filter.ProvideAccessLogFilter,
		filter.ProvideAuthFilter,
		router.WireSet,
	))
}
