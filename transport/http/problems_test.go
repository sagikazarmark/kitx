package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/moogar0880/problems"
)

type errorMatcherStub struct {
	match bool
}

func (e errorMatcherStub) MatchError(err error) bool {
	return e.match
}

func TestNewStatusProblemMatcher(t *testing.T) {
	matcher := NewStatusProblemMatcher(http.StatusNotFound, errorMatcherStub{true})

	if !matcher.MatchError(errors.New("error")) {
		t.Error("error is supposed to be matched")
	}

	if want, have := http.StatusNotFound, matcher.Status(); want != have {
		t.Errorf("unexpected status\nexpected: %d\nactual:   %d", want, have)
	}
}

type matcherStub struct {
	err error
}

func (s matcherStub) MatchError(err error) bool {
	return s.err == err
}

type matcherFactoryStub struct {
	err error
}

func (s matcherFactoryStub) MatchError(err error) bool {
	return s.err == err
}

func (s matcherFactoryStub) NewProblem(_ context.Context, err error) problems.Problem {
	return problems.NewDetailedProblem(http.StatusServiceUnavailable, "my error")
}

type statusMatcherStub struct {
	err    error
	status int
}

func (s statusMatcherStub) MatchError(err error) bool {
	return s.err == err
}

func (s statusMatcherStub) Status() int {
	return s.status
}

type statusMatcherFactoryStub struct {
	statusMatcherStub
}

func (s statusMatcherFactoryStub) NewProblem(_ context.Context, _ error) problems.Problem {
	return problems.NewDetailedProblem(http.StatusBadRequest, "custom error")
}

type statusMatcherStatusFactoryStub struct {
	err    error
	status int
}

func (s statusMatcherStatusFactoryStub) MatchError(err error) bool {
	return s.err == err
}

func (s statusMatcherStatusFactoryStub) Status() int {
	return s.status
}

func (s statusMatcherStatusFactoryStub) NewStatusProblem(
	_ context.Context,
	status int,
	_ error,
) problems.StatusProblem {
	return problems.NewDetailedProblem(status, "custom status error")
}

func testProblemEquals(t *testing.T, problem *problems.DefaultProblem, status int, detail string) {
	t.Helper()

	if want, have := status, problem.Status; want != have {
		t.Errorf("unexpected status\nexpected: %d\nactual:   %d", want, have)
	}

	if want, have := detail, problem.Detail; want != have {
		t.Errorf("unexpected status\nexpected: %s\nactual:   %s", want, have)
	}
}

func TestProblemFactory(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		tests := []ProblemFactory{
			NewDefaultProblemFactory(),
			NewProblemFactory(ProblemFactoryConfig{}),
		}

		for _, factory := range tests {
			factory := factory

			t.Run("", func(t *testing.T) {
				problem := factory.NewProblem(context.Background(), errors.New("error")).(*problems.DefaultProblem)

				testProblemEquals(t, problem, http.StatusInternalServerError, "something went wrong")
			})
		}
	})

	t.Run("matcher", func(t *testing.T) {
		err := errors.New("error")

		tests := []struct {
			config ProblemFactoryConfig
			status int
			detail string
		}{
			{
				config: ProblemFactoryConfig{
					Matchers: []ProblemMatcher{
						statusMatcherStub{
							err:    err,
							status: http.StatusNotFound,
						},
					},
				},
				status: http.StatusNotFound,
				detail: "error",
			},
			{
				config: ProblemFactoryConfig{
					Matchers: []ProblemMatcher{
						statusMatcherFactoryStub{
							statusMatcherStub: statusMatcherStub{
								err:    err,
								status: http.StatusNotFound,
							},
						},
					},
				},
				status: http.StatusBadRequest,
				detail: "custom error",
			},
			{
				config: ProblemFactoryConfig{
					Matchers: []ProblemMatcher{
						statusMatcherStatusFactoryStub{
							err:    err,
							status: http.StatusNotFound,
						},
					},
				},
				status: http.StatusNotFound,
				detail: "custom status error",
			},
			{
				config: ProblemFactoryConfig{
					Matchers: []ProblemMatcher{
						matcherStub{
							err: err,
						},
					},
				},
				status: http.StatusInternalServerError,
				detail: "error",
			},
			{
				config: ProblemFactoryConfig{
					Matchers: []ProblemMatcher{
						matcherFactoryStub{
							err: err,
						},
					},
				},
				status: http.StatusServiceUnavailable,
				detail: "my error",
			},
		}

		for _, test := range tests {
			test := test

			t.Run("", func(t *testing.T) {
				factory := NewProblemFactory(test.config)

				problem := factory.NewProblem(context.Background(), err).(*problems.DefaultProblem)

				testProblemEquals(t, problem, test.status, test.detail)
			})
		}
	})
}

func ExampleNewProblemFactory() {
	factory := NewProblemFactory(ProblemFactoryConfig{
		Matchers: []ProblemMatcher{
			NewStatusProblemMatcher(http.StatusNotFound, ErrorMatcherFunc(func(err error) bool {
				return err.Error() == "not found"
			})),
		},
	})

	err := errors.New("not found")

	problem := factory.NewProblem(context.Background(), err).(*problems.DefaultProblem)

	fmt.Println(problem.Status, problem.Detail)

	// Output: 404 not found
}
