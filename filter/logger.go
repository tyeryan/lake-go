package filter

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/middleware"
	logutil "github.com/tyeryan/l-protocol/log"
	"io/ioutil"
	"net/http"
)

// RequestResponseLogger returns a logger handler which logs http request and response.
func RequestResponseLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			log := logutil.GetLogger("httpLogger")

			log.Add("method", r.Method)
			log.Add("remote", r.RemoteAddr)
			log.Add("proto", r.Proto)

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			log.Add("path", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI))

			if r.GetBody != nil {
				body, err := r.GetBody()
				if err == nil {
					reqBody, err := ioutil.ReadAll(body)
					if err == nil {
						log.Add("reqBody", string(reqBody))
					}
				}
			}

			log.Infow(r.Context(), "http request")

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			var rspWriter bytes.Buffer
			ww.Tee(&rspWriter)

			defer func() {
				log.Add("httpStatus", ww.Status())
				log.Add("rspBody", rspWriter)
				log.Canonical(r.Context(), "http response", nil, recover())
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
