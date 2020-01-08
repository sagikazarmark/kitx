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

func (e errorMatcherStub) MatchError(_ error) bool {
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

type matcherConverterStub struct {
	err error
}

func (s matcherConverterStub) MatchError(err error) bool {
	return s.err == err
}

func (s matcherConverterStub) NewProblem(_ context.Context, _ error) problems.Problem {
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

type statusMatcherConverterStub struct {
	statusMatcherStub
}

func (s statusMatcherConverterStub) NewProblem(_ context.Context, _ error) problems.Problem {
	return problems.NewDetailedProblem(http.StatusBadRequest, "custom error")
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

func TestProblemConverter(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		tests := []ProblemConverter{
			NewDefaultProblemConverter(),
			NewProblemConverter(ProblemConverterConfig{}),
		}

		for _, problemConverter := range tests {
			problemConverter := problemConverter

			t.Run("", func(t *testing.T) {
				problem := problemConverter.NewProblem(context.Background(), errors.New("error")).(*problems.DefaultProblem)

				testProblemEquals(t, problem, http.StatusInternalServerError, "something went wrong")
			})
		}
	})

	t.Run("matcher", func(t *testing.T) {
		err := errors.New("error")

		tests := []struct {
			config ProblemConverterConfig
			status int
			detail string
		}{
			{
				config: ProblemConverterConfig{
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
				config: ProblemConverterConfig{
					Matchers: []ProblemMatcher{
						statusMatcherConverterStub{
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
				config: ProblemConverterConfig{
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
				config: ProblemConverterConfig{
					Matchers: []ProblemMatcher{
						matcherConverterStub{
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
				problemConverter := NewProblemConverter(test.config)

				problem := problemConverter.NewProblem(context.Background(), err).(*problems.DefaultProblem)

				testProblemEquals(t, problem, test.status, test.detail)
			})
		}
	})
}

func ExampleNewProblemConverter() {
	problemConverter := NewProblemConverter(ProblemConverterConfig{
		Matchers: []ProblemMatcher{
			NewStatusProblemMatcher(http.StatusNotFound, ErrorMatcherFunc(func(err error) bool {
				return err.Error() == "not found"
			})),
		},
	})

	err := errors.New("not found")

	problem := problemConverter.NewProblem(context.Background(), err).(*problems.DefaultProblem)

	fmt.Println(problem.Status, problem.Detail)

	// Output: 404 not found
}
