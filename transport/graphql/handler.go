package graphql

import (
	"context"
)

type errorEncoderHandler struct {
	handler      Handler
	errorEncoder EncodeErrorResponseFunc
}

// NewErrorEncoderHandler wraps a GraphQL handler and encodes the returned error using the error encoder (if necessary).
//
// For example, a returned endpoint error might need additional encoding.
func NewErrorEncoderHandler(handler Handler, errorEncoder EncodeErrorResponseFunc) Handler {
	return errorEncoderHandler{
		handler:      handler,
		errorEncoder: errorEncoder,
	}
}

func (h errorEncoderHandler) ServeGraphQL(ctx context.Context, req interface{}) (context.Context, interface{}, error) {
	ctx, resp, err := h.handler.ServeGraphQL(ctx, req)
	if err != nil {
		return ctx, resp, h.errorEncoder(ctx, err)
	}

	return ctx, resp, nil
}
