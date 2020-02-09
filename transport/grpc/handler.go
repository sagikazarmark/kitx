package grpc

import (
	"context"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

type errorEncoderHandler struct {
	handler      kitgrpc.Handler
	errorEncoder EncodeErrorResponseFunc
}

// NewErrorEncoderHandler wraps a gRPC handler and encodes the returned error using the error encoder (if necessary).
//
// For example, a returned endpoint error might need additional encoding.
func NewErrorEncoderHandler(handler kitgrpc.Handler, errorEncoder EncodeErrorResponseFunc) kitgrpc.Handler {
	return errorEncoderHandler{
		handler:      handler,
		errorEncoder: errorEncoder,
	}
}

func (h errorEncoderHandler) ServeGRPC(ctx context.Context, req interface{}) (context.Context, interface{}, error) {
	ctx, resp, err := h.handler.ServeGRPC(ctx, req)
	if err != nil {
		return ctx, resp, h.errorEncoder(ctx, err)
	}

	return ctx, resp, nil
}
