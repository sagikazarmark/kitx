package endpoint

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-kit/kit/endpoint"
)

// nolint: gochecknoglobals
var (
	ctx = context.Background()
	req = struct{}{}
)

func ExampleChain() {
	annotate := func(pos string) endpoint.Middleware {
		return func(e endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req interface{}) (response interface{}, err error) {
				fmt.Println(pos + " pre")

				response, err = e(ctx, req)

				fmt.Println(pos + " post")

				return
			}
		}
	}
	e := endpoint.Chain(
		annotate("first"),
		annotate("second"),
		annotate("third"),
	)(
		func(context.Context, interface{}) (interface{}, error) {
			fmt.Println("endpoint")

			return nil, nil
		},
	)

	if _, err := e(ctx, req); err != nil {
		panic(err)
	}

	// Output:
	// first pre
	// second pre
	// third pre
	// endpoint
	// third post
	// second post
	// first post
}

func TestFailerMiddleware(t *testing.T) {
	var endpointCalled bool
	berr := errors.New("error")

	var e endpoint.Endpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		endpointCalled = true

		return nil, berr
	}

	e = FailerMiddleware(ErrorMatcherFunc(func(err error) bool {
		return true
	}))(e)

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
			t.Error("failure response is supposed to return the wrapped error")
		}
	}
}
