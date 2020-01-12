package transport

import (
	"context"

	"github.com/go-kit/kit/transport"
)

// ErrorHandler is a generic error handler that allows applications (and libraries) to handle errors
// without worrying about the actual error handling strategy (logging, error tracking service, etc).
//
// Compared to go-kit's error handler, it does not accept a context as the first parameter.
//
// ErrorHandler is derived from https://godoc.org/emperror.dev/emperror#Handler
type ErrorHandler interface {
	Handle(err error)
}

// ErrorHandlerContext is an optional interface that MAY be implemented by an ErrorHandler.
// It is similar to ErrorHandler, but it receives a context as the first parameter.
// An implementation MAY extract information from the context and annotate err with it.
//
// This interface is closer to go-kit's error handler, but the method name doesn't match,
// so it needs to be wrapped in a compatibility layer.
//
// ErrorHandlerContext MAY honor the deadline carried by the context, but that's not a hard requirement.
//
// ErrorHandlerContext is derived from https://godoc.org/emperror.dev/emperror#ContextAwareHandler
type ErrorHandlerContext interface {
	HandleContext(ctx context.Context, err error)
}

// NewErrorHandler turns an ErrorHandler (or an ErrorHandlerContext) into a go-kit compatible error handler.
func NewErrorHandler(handler ErrorHandler) transport.ErrorHandler {
	if handler, ok := handler.(ErrorHandlerContext); ok {
		return errorHandlerContext{handler}
	}

	return errorHandler{handler}
}

type errorHandler struct {
	handler ErrorHandler
}

func (e errorHandler) Handle(_ context.Context, err error) {
	e.handler.Handle(err)
}

type errorHandlerContext struct {
	handler ErrorHandlerContext
}

func (e errorHandlerContext) Handle(ctx context.Context, err error) {
	e.handler.HandleContext(ctx, err)
}
