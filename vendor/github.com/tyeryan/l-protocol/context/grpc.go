package ctxutil

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// Add add kv context
func Add(ctx context.Context, key ContextKey, value string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, string(key), value)
}

// Read read val from context
func Read(ctx context.Context, key ContextKey) (string, bool) {
	//why read outgoing first
	//reason: plz read common-component.grpcserver.request_context_interceptor.go, it auto copy the incoming meta data to out going
	outMD, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		val, mdOK := outMD[string(key)]
		if mdOK {
			if len(val) > 0 {
				//len(val) - 1 => we want the latest value
				return val[(len(val) - 1)], true
			}
		}
	}

	inMD, ok := metadata.FromIncomingContext(ctx)
	if ok {
		val, mdOK := inMD[string(key)]
		if mdOK {
			//len(val) - 1 => we want the latest value
			if len(val) > 0 {
				return val[(len(val) - 1)], true
			}
		}
	}

	return "", false
}
