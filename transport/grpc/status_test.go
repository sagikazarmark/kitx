package grpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errorMatcherStub struct {
	match bool
}

func (e errorMatcherStub) MatchError(_ error) bool {
	return e.match
}

func TestNewStatusCodeMatcher(t *testing.T) {
	matcher := NewStatusCodeMatcher(codes.NotFound, errorMatcherStub{true})

	if !matcher.MatchError(errors.New("error")) {
		t.Error("error is supposed to be matched")
	}

	if want, have := codes.NotFound, matcher.Code(); want != have {
		t.Errorf("unexpected code\nexpected: %d\nactual:   %d", want, have)
	}
}

type matcherStub struct {
	err error
}

func (s matcherStub) MatchError(err error) bool {
	return s.err == err
}

type matcherConverterStub struct {
	err error
}

func (s matcherConverterStub) MatchError(err error) bool {
	return s.err == err
}

func (s matcherConverterStub) NewStatus(_ context.Context, _ error) *status.Status {
	return status.New(codes.Unavailable, "my error")
}

type statusMatcherStub struct {
	err  error
	code codes.Code
}

func (s statusMatcherStub) MatchError(err error) bool {
	return s.err == err
}

func (s statusMatcherStub) Code() codes.Code {
	return s.code
}

type statusMatcherConverterStub struct {
	statusMatcherStub
}

func (s statusMatcherConverterStub) NewStatus(_ context.Context, _ error) *status.Status {
	return status.New(codes.InvalidArgument, "custom error")
}

func testStatusEquals(t *testing.T, s *status.Status, code codes.Code, message string) {
	t.Helper()

	if want, have := code, s.Code(); want != have {
		t.Errorf("unexpected code\nexpected: %d\nactual:   %d", want, have)
	}

	if want, have := message, s.Message(); want != have {
		t.Errorf("unexpected code\nexpected: %s\nactual:   %s", want, have)
	}
}

func TestStatusConverter(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		tests := []StatusConverter{
			NewDefaultStatusConverter(),
			NewStatusConverter(StatusConverterConfig{}),
		}

		for _, statusConverter := range tests {
			statusConverter := statusConverter

			t.Run("", func(t *testing.T) {
				s := statusConverter.NewStatus(context.Background(), errors.New("error"))

				testStatusEquals(t, s, codes.Internal, "something went wrong")
			})
		}
	})

	t.Run("matcher", func(t *testing.T) {
		err := errors.New("error")

		tests := []struct {
			config  StatusConverterConfig
			code    codes.Code
			message string
		}{
			{
				config: StatusConverterConfig{
					Matchers: []StatusMatcher{
						statusMatcherStub{
							err:  err,
							code: codes.NotFound,
						},
					},
				},
				code:    codes.NotFound,
				message: "error",
			},
			{
				config: StatusConverterConfig{
					Matchers: []StatusMatcher{
						statusMatcherConverterStub{
							statusMatcherStub: statusMatcherStub{
								err:  err,
								code: http.StatusNotFound,
							},
						},
					},
				},
				code:    codes.InvalidArgument,
				message: "custom error",
			},
			{
				config: StatusConverterConfig{
					Matchers: []StatusMatcher{
						matcherStub{
							err: err,
						},
					},
				},
				code:    codes.Internal,
				message: "error",
			},
			{
				config: StatusConverterConfig{
					Matchers: []StatusMatcher{
						matcherConverterStub{
							err: err,
						},
					},
				},
				code:    codes.Unavailable,
				message: "my error",
			},
		}

		for _, test := range tests {
			test := test

			t.Run("", func(t *testing.T) {
				statusConverter := NewStatusConverter(test.config)

				s := statusConverter.NewStatus(context.Background(), err)

				testStatusEquals(t, s, test.code, test.message)
			})
		}
	})
}

func ExampleNewStatusConverter() {
	statusConverter := NewStatusConverter(StatusConverterConfig{
		Matchers: []StatusMatcher{
			NewStatusCodeMatcher(codes.NotFound, ErrorMatcherFunc(func(err error) bool {
				return err.Error() == "not found"
			})),
		},
	})

	err := errors.New("not found")

	s := statusConverter.NewStatus(context.Background(), err)

	fmt.Println(s.Code(), s.Message())

	// Output: NotFound not found
}
