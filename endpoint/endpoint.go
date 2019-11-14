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

// ErrorMatcher is a predicate for errors.
// It can be used in middleware to decide whether to take action or not.
type ErrorMatcher interface {
	// MatchError evaluates the predicate for an error.
	MatchError(err error) bool
}

// ErrorMatcherFunc turns a plain function into an ErrorMatcher if it's definition matches the interface.
type ErrorMatcherFunc func(err error) bool

// MatchError calls the underlying function to evaluate the predicate.
func (fn ErrorMatcherFunc) MatchError(err error) bool {
	return fn(err)
}

type failer struct {
	err error
}

func (f failer) Failed() error {
	return f.err
}

// FailerMiddleware checks if a returned error matches a predicate and wraps it in a failer response if it does.
func FailerMiddleware(matcher ErrorMatcher) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			resp, err := e(ctx, request)
			if err != nil && matcher.MatchError(err) {
				return failer{err}, nil
			}

			return resp, err
		}
	}
}

type businessError interface {
	// IsBusinessError checks if an error should be returned as a business error from an endpoint.
	IsBusinessError() bool
}

// BusinessErrorMiddleware checks if a returned error is a business error and wraps it in a failer response if it is.
// Deprecated: Use FailerMiddleware instead.
func BusinessErrorMiddleware(e endpoint.Endpoint) endpoint.Endpoint {
	return FailerMiddleware(ErrorMatcherFunc(func(err error) bool {
		var berr businessError

		return errors.As(err, &berr) && berr.IsBusinessError()
	}))(e)
}
