package grpc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EncodeErrorResponseFunc transforms the passed error to a gRPC code error.
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

// StatusConverter creates a new gRPC Status from an error.
type StatusConverter interface {
	// NewStatus creates a new gRPC Status from an error.
	NewStatus(ctx context.Context, err error) *status.Status
}

type defaultErrorStatusConverter struct{}

func (d defaultErrorStatusConverter) NewStatus(_ context.Context, _ error) *status.Status {
	return status.New(codes.Internal, "something went wrong")
}

// NewStatusErrorResponseEncoder returns an error response encoder that encodes errors as gRPC Status errors.
func NewStatusErrorResponseEncoder(statusConverter StatusConverter) EncodeErrorResponseFunc {
	return func(ctx context.Context, err error) error {
		// Do not convert gRPC errors
		if IsGRPCError(err) {
			return err
		}

		return statusConverter.NewStatus(ctx, err).Err()
	}
}

// NewDefaultStatusErrorResponseEncoder returns an error response encoder that encodes errors as gRPC Status errors.
//
// The returned encoder encodes every error as Internal error.
func NewDefaultStatusErrorResponseEncoder() EncodeErrorResponseFunc {
	return NewStatusErrorResponseEncoder(defaultErrorStatusConverter{})
}
