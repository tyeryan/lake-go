package router

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/google/wire"
	"github.com/tyeryan/l-common-util/apm"
	"lake-go/filter"
	"lake-go/grpcclient"
	"lake-go/handler/auth"
	"net/http"
	"time"
)

var (
	WireSet = wire.NewSet(
		ProvideRoutes,
		ProvideLakeHandler,
		grpcclient.ProvideLAuthClient,
		grpcclient.ProvideLAuthConfig,
		auth.ProvideAuthHandler,
	)
)

type UserIDContextKey string

func ProvideRoutes(
	authFilter *filter.AuthFilter,
	lakeHandler *LakeHandler,
	authHandler *auth.AuthHandler,
	apmConfig *apm.ApmConfig,
	accessLogFilter *filter.AccessLogFilter,
) http.Handler {
	r := chi.NewRouter()

	//r.Use(middleware.RealIP)
	//if apmConfig.Enable {
	//	r.Use(apmc)
	//}

	r.Use(filter.RequestResponseLogger())
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Goog-AuthUser", "X-Request-Id"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/v1", func(r chi.Router) {
		r.Use(accessLogFilter.Filter())
		r.Get("/healthcheck", lakeHandler.HealthCheck)
	})

	return r
}
