package ctxutil

var (
	defaultOptions = options{}
)

type options struct {
	stan string
}

// ContextOption context options
type ContextOption func(*options)

// WithStan pre-set the stan value in context
func WithStan(stan string) ContextOption {
	return func(o *options) {
		o.stan = stan
	}
}
