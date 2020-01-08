package http

import (
	"context"
	"net/http"

	"emperror.dev/errors"
	"github.com/moogar0880/problems"
)

// ErrorMatcher checks if an error matches a predefined set of conditions.
type ErrorMatcher interface {
	// MatchError evaluates the predefined set of conditions for err.
	MatchError(err error) bool
}

// ErrorMatcherFunc turns a plain function into an ErrorMatcher if it's definition matches the interface.
type ErrorMatcherFunc func(err error) bool

// MatchError calls the underlying function to check if err matches a certain condition.
func (fn ErrorMatcherFunc) MatchError(err error) bool {
	return fn(err)
}

// ProblemMatcher matches an error.
// It is an alias to the ErrorMatcher interface.
type ProblemMatcher interface {
	ErrorMatcher
}

// StatusProblemMatcher matches an error and returns the appropriate status code for it.
type StatusProblemMatcher interface {
	ProblemMatcher

	// Status returns the HTTP status code.
	Status() int
}

type statusProblemMatcher struct {
	status  int
	matcher ErrorMatcher
}

// NewStatusProblemMatcher returns a new StatusProblemMatcher.
func NewStatusProblemMatcher(status int, matcher ProblemMatcher) StatusProblemMatcher {
	return statusProblemMatcher{
		status:  status,
		matcher: matcher,
	}
}

func (m statusProblemMatcher) MatchError(err error) bool {
	return m.matcher.MatchError(err)
}

func (m statusProblemMatcher) Status() int {
	return m.status
}

// StatusProblemConverter creates a new status problem instance.
type StatusProblemConverter interface {
	// NewStatusProblem creates a new status problem instance.
	NewStatusProblem(ctx context.Context, status int, err error) problems.StatusProblem
}

type defaultProblemConverter struct{}

func (c defaultProblemConverter) NewProblem(_ context.Context, err error) problems.Problem {
	return problems.NewDetailedProblem(http.StatusInternalServerError, err.Error())
}

func (c defaultProblemConverter) NewStatusProblem(_ context.Context, status int, err error) problems.StatusProblem {
	return problems.NewDetailedProblem(status, err.Error())
}

type problemConverter struct {
	matchers []ProblemMatcher

	problemConverter       ProblemConverter
	statusProblemConverter StatusProblemConverter
}

// ProblemConverterConfig configures the ProblemConverter implementation.
type ProblemConverterConfig struct {
	// Matchers are used to match errors and create problems.
	// By default an empty detailed problem is created.
	// If no matchers match the error (or no matchers are configured) a fallback problem is created/returned.
	//
	// If a matcher also implements ProblemConverter it is used instead of the builtin ProblemConverter
	// for creating the problem instance.
	//
	// If a matchers also implements StatusProblemMatcher and StatusProblemConverter
	// it is used instead of the builtin StatusProblemConverter for creating the problem instance.
	//
	// If a matchers also implements StatusProblemMatcher (but not StatusProblemConverter)
	// the builtin StatusProblemConverter is used for creating the problem instance.
	Matchers []ProblemMatcher

	// Problem converters used for converting errors to problems.
	ProblemConverter       ProblemConverter
	StatusProblemConverter StatusProblemConverter
}

// NewProblemConverter returns a new ProblemConverter implementation.
func NewProblemConverter(config ProblemConverterConfig) ProblemConverter {
	c := problemConverter{
		matchers:               config.Matchers,
		problemConverter:       config.ProblemConverter,
		statusProblemConverter: config.StatusProblemConverter,
	}

	if c.problemConverter == nil {
		c.problemConverter = defaultProblemConverter{}
	}

	if c.statusProblemConverter == nil {
		if spc, ok := c.problemConverter.(StatusProblemConverter); ok {
			c.statusProblemConverter = spc
		} else {
			c.statusProblemConverter = defaultProblemConverter{}
		}
	}

	return c
}

// NewDefaultProblemConverter returns a new ProblemConverter implementation with default configuration.
func NewDefaultProblemConverter() ProblemConverter {
	return NewProblemConverter(ProblemConverterConfig{})
}

func (c problemConverter) NewProblem(ctx context.Context, err error) problems.Problem {
	for _, matcher := range c.matchers {
		if matcher.MatchError(err) {
			if converter, ok := matcher.(ProblemConverter); ok {
				return converter.NewProblem(ctx, err)
			}

			if statusMatcher, ok := matcher.(StatusProblemMatcher); ok {
				if converter, ok := statusMatcher.(StatusProblemConverter); ok {
					return converter.NewStatusProblem(ctx, statusMatcher.Status(), err)
				}

				return c.statusProblemConverter.NewStatusProblem(ctx, statusMatcher.Status(), err)
			}

			return c.problemConverter.NewProblem(ctx, err)
		}
	}

	return c.statusProblemConverter.NewStatusProblem(
		ctx,
		http.StatusInternalServerError,
		errors.New("something went wrong"),
	)
}
