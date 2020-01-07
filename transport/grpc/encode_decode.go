package grpc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

// EncodeErrorResponseFunc transforms the passed error to a gRPC status error.
// It's designed to be used in gRPC servers, for server-side endpoints.
type EncodeErrorResponseFunc func(context.Context, error) error

// ErrorResponseEncoder encodes the passed response object to a gRPC response or error.
func ErrorResponseEncoder(
	encoder kitgrpc.EncodeResponseFunc,
	errorEncoder EncodeErrorResponseFunc,
) kitgrpc.EncodeResponseFunc {
	return func(ctx context.Context, resp interface{}) (interface{}, error) {
		if f, ok := resp.(endpoint.Failer); ok && f.Failed() != nil {
			return nil, errorEncoder(ctx, f.Failed())
		}

		return encoder(ctx, resp)
	}
}
