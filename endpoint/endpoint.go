package endpoint

import (
	"context"

	"emperror.dev/errors"
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

type businessError interface {
	// IsBusinessError checks if an error should be returned as a business error from an endpoint.
	IsBusinessError() bool
}

type failer struct {
	err error
}

func (f failer) Failed() error {
	return f.err
}

// BusinessErrorMiddleware checks if a returned error is a business error and wraps it in a failer response if it is.
func BusinessErrorMiddleware(e endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		resp, err := e(ctx, request)
		if err != nil {
			var berr businessError
			if errors.As(err, &berr) && berr.IsBusinessError() {
				return failer{err}, nil
			}
		}

		return resp, err
	}
}
