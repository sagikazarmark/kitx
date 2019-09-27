package endpoint

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kit/kit/endpoint"
)

type bError struct {
	err string
}

func (b bError) Error() string {
	return b.err
}

func (b bError) IsBusinessError() bool {
	return true
}

func TestBusinessErrorMiddleware(t *testing.T) {
	var endpointCalled bool
	berr := bError{"error"}

	var e endpoint.Endpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		endpointCalled = true

		return nil, berr
	}

	e = BusinessErrorMiddleware(e)

	resp, err := e(context.Background(), nil)

	if !endpointCalled {
		t.Error("endpoint is supposed to be called")
	}

	if err != nil {
		t.Error("error is supposed to be wrapped by the response")
	}

	if failer, ok := resp.(endpoint.Failer); !ok {
		t.Error("response is supposed to be a failure response")

		if !errors.Is(failer.Failed(), berr) {
			t.Error("failure response is supposed to return the business error")
		}
	}
}
