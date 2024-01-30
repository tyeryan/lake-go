package ctxutil

import (
	"context"
)

// ContextKey meta data key in context
type ContextKey string

const (
	//Stan stan, same as the request id from client, or generated from system
	Stan ContextKey = "x-stan-id"
	//UserID very clear la
	UserID ContextKey = "x-user-id"
	//BasicAuthKey basic auth key
	BasicAuthKey ContextKey = "x-basic-auth-key"
	//BasicAuthSecret basic auth secret
	BasicAuthSecret ContextKey = "x-basic-auth-secret"

	//For mock testing
	YouMetadata ContextKey = "x-you-metadata"

	YBPCurrentOperator ContextKey = "x-ybp-current-operator"
)

// NewContext create new context
func NewContext(opt ...ContextOption) context.Context {
	currentCtx := context.TODO()

	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}

	if opts.stan != "" {
		currentCtx = Add(currentCtx, Stan, opts.stan)
	}

	return currentCtx
}
