package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Chain composes a single middleware from a list.
// Compared to endpoint.Chain, this function accepts a variadic list.
func Chain(mw ...endpoint.Middleware) func(endpoint.Endpoint) endpoint.Endpoint {
	if len(mw) == 0 {
		return func(e endpoint.Endpoint) endpoint.Endpoint {
			return e
		}
	}

	return func(e endpoint.Endpoint) endpoint.Endpoint {
		for i := len(mw) - 1; i >= 0; i-- { // traverse middleware in a reverse order
			e = mw[i](e)
		}

		return e
	}
}

// ErrorMatcher is a predicate for errors.
// It can be used in middleware to decide whether to take action or not.
type ErrorMatcher func(err error) bool

type failer struct {
	err error
}

func (f failer) Failed() error {
	return f.err
}

// FailerMiddleware checks if a returned error matches a predicate and wraps it in a failer response if it does.
func FailerMiddleware(errorMatcher ErrorMatcher) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			resp, err := e(ctx, request)
			if err != nil && errorMatcher(err) {
				return failer{err}, nil
			}

			return resp, err
		}
	}
}

type contextKey string

// operationNameContextKey holds the key used to store an operation name in the context.
const operationNameContextKey contextKey = "operationName"

// OperationNameMiddleware populates the context with a common name for the endpoint.
// It can be used in subsequent endpoints in the chain to identify the operation.
func OperationNameMiddleware(name string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = context.WithValue(ctx, operationNameContextKey, name)

			return next(ctx, req)
		}
	}
}

// OperationName fetches the endpoint operation name from the context (if any).
// If an endpoint name is not found or it isn't string, the second return argument is false.
func OperationName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(operationNameContextKey).(string)

	return name, ok
}
