package endpoint

import (
	"context"

	"emperror.dev/errors"
	"github.com/go-kit/kit/endpoint"
)

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
