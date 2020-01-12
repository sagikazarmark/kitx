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
	e := Chain(
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

func ExampleCombine() {
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
	e := Combine(
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
	t.Run("failed", func(t *testing.T) {
		var endpointCalled bool
		berr := errors.New("error")

		var e endpoint.Endpoint = func(ctx context.Context, request interface{}) (interface{}, error) {
			endpointCalled = true

			return nil, berr
		}

		e = FailerMiddleware(func(err error) bool { return err != nil })(e)

		resp, err := e(context.Background(), nil)

		if !endpointCalled {
			t.Fatal("endpoint is supposed to be called")
		}

		if err != nil {
			t.Fatal("error is supposed to be wrapped by the response")
		}

		if failer, ok := resp.(endpoint.Failer); !ok {
			t.Error("response is supposed to be a failure response")

			if !errors.Is(failer.Failed(), berr) {
				t.Error("failure response is supposed to return the wrapped error")
			}
		}
	})

	t.Run("success", func(t *testing.T) {
		var endpointCalled bool
		const response = "response"

		var e endpoint.Endpoint = func(ctx context.Context, request interface{}) (interface{}, error) {
			endpointCalled = true

			return response, nil
		}

		e = FailerMiddleware(func(err error) bool { return err != nil })(e)

		resp, err := e(context.Background(), nil)

		if !endpointCalled {
			t.Fatal("endpoint is supposed to be called")
		}

		if err != nil {
			t.Fatal("unexpected error: ", err)
		}

		if resp != response {
			t.Errorf("unexpected response\nexpected: %s\nactual:   %s", response, resp)
		}
	})
}

func TestOperationNameMiddleware(t *testing.T) {
	ctx := context.Background()

	var name string

	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		name, _ = OperationName(ctx)

		return nil, nil
	}

	mw := OperationNameMiddleware("go-kit/endpoint")

	_, _ = mw(ep)(ctx, nil)

	if want, have := "go-kit/endpoint", name; want != have {
		t.Fatalf("unexpected endpoint name, wanted %q, got %q", want, have)
	}
}
