package graphql

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// DecodeRequestFunc extracts a user-domain request object from a GraphQL request.
// It's designed to be used in GraphQL servers, for server-side endpoints. One
// straightforward DecodeRequestFunc could be something that decodes from the
// GraphQL request message to the concrete request type.
type DecodeRequestFunc func(context.Context, interface{}) (request interface{}, err error)

// EncodeResponseFunc encodes the passed response object to the GraphQL response
// message. It's designed to be used in GraphQL servers, for server-side endpoints.
// One straightforward EncodeResponseFunc could be something that encodes the
// object directly to the GraphQL response message.
type EncodeResponseFunc func(context.Context, interface{}) (response interface{}, err error)

// EncodeErrorResponseFunc transforms the passed error to a GraphQL error.
// It's designed to be used in GraphQL servers, for server-side endpoints.
type EncodeErrorResponseFunc func(context.Context, error) error

// ErrorResponseEncoder encodes the passed response object to a GraphQL response or error.
func ErrorResponseEncoder(
	encoder EncodeResponseFunc,
	errorEncoder EncodeErrorResponseFunc,
) EncodeResponseFunc {
	return func(ctx context.Context, resp interface{}) (interface{}, error) {
		if f, ok := resp.(endpoint.Failer); ok && f.Failed() != nil {
			return nil, errorEncoder(ctx, f.Failed())
		}

		return encoder(ctx, resp)
	}
}
