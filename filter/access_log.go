package filter

import (
	"github.com/google/wire"
	lapm "github.com/tyeryan/l-common-util/apm"
	ctxutil "github.com/tyeryan/l-protocol/context"
	"go.elastic.co/apm"
	"net/http"
)

var (
	WireSet = wire.NewSet(
		ProvideAccessLogFilter,
	)
)

const UserReferenceID ctxutil.ContextKey = "x-user-reference-id"

func ProvideAccessLogFilter(apmConfig *lapm.ApmConfig) *AccessLogFilter {
	return &AccessLogFilter{
		apmConfig: apmConfig,
	}
}

type AccessLogFilter struct {
	apmConfig *lapm.ApmConfig
}

// Be aware to include this filter after access_token filter
func (f *AccessLogFilter) Filter() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if f.apmConfig.Enable {
				if userID, ok := ctxutil.Read(r.Context(), UserReferenceID); ok {
					tx := apm.TransactionFromContext(r.Context())
					tx.Context.SetUserID(userID)
					tx.Context.SetLabel("api-server", "c-user-api")
				}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
