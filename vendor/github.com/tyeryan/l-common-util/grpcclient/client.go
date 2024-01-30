package grpcclient

import (
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.elastic.co/apm/module/apmgrpc"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type options struct {
	timeout      time.Duration
	retryBackoff time.Duration
}

var messageSizeLimit = 30 * 1024 * 1024

var defaultOptions = options{
	timeout:      time.Duration(30 * time.Second),
	retryBackoff: time.Duration(1 * time.Second),
}

type Option func(*options)

func WithTimeout(t time.Duration) Option {
	return func(o *options) {
		o.timeout = t
	}
}

func WithRetryBackoff(t time.Duration) Option {
	return func(o *options) {
		o.retryBackoff = t
	}
}

func NewGRPCConnection(address string, opts ...Option) (*grpc.ClientConn, error) {
	options := defaultOptions
	for _, o := range opts {
		o(&options)
	}

	return grpc.Dial(address,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(messageSizeLimit),
			grpc.MaxCallSendMsgSize(messageSizeLimit)),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"`),
		grpc.WithTimeout(options.timeout),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(options.retryBackoff)),
			grpc_retry.WithCodes(codes.Internal, codes.Unavailable, codes.Unknown, codes.Unimplemented),
		)),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				apmgrpc.NewUnaryClientInterceptor(),
				grpc_retry.UnaryClientInterceptor(
					grpc_retry.WithBackoff(grpc_retry.BackoffLinear(options.retryBackoff)),
					grpc_retry.WithCodes(codes.Internal, codes.Unavailable, codes.Unknown, codes.Unimplemented),
				))))
}
