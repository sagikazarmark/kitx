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

// StatusProblemFactory creates a new status problem instance.
type StatusProblemFactory interface {
	// NewStatusProblem creates a new status problem instance.
	NewStatusProblem(ctx context.Context, status int, err error) problems.StatusProblem
}

type defaultProblemFactory struct{}

func (d defaultProblemFactory) NewProblem(_ context.Context, err error) problems.Problem {
	return problems.NewDetailedProblem(http.StatusInternalServerError, err.Error())
}

func (d defaultProblemFactory) NewStatusProblem(_ context.Context, status int, err error) problems.StatusProblem {
	return problems.NewDetailedProblem(status, err.Error())
}

type problemFactory struct {
	matchers []ProblemMatcher

	problemFactory       ProblemFactory
	statusProblemFactory StatusProblemFactory

	fallbackProblem problems.Problem
}

// ProblemFactoryConfig configures the ProblemFactory implementation.
type ProblemFactoryConfig struct {
	// Matchers are used to match errors and create problems.
	// By default an empty detailed problem is created.
	// If no matchers match the error (or no matchers are configured) a fallback problem is created/returned.
	//
	// If a matcher also implements ProblemFactory it is used instead of the builtin ProblemFactory
	// for creating the problem instance.
	//
	// If a matchers also implements StatusProblemMatcher and StatusProblemFactory
	// it is used instead of the builtin StatusProblemFactory for creating the problem instance.
	//
	// If a matchers also implements StatusProblemMatcher (but not StatusProblemFactory)
	// the builtin StatusProblemFactory is used for creating the problem instance.
	Matchers []ProblemMatcher

	// Problem factories used for creating problems.
	ProblemFactory       ProblemFactory
	StatusProblemFactory StatusProblemFactory

	// FallbackProblem is an optional problem instance returned when not matchers match the error.
	FallbackProblem problems.Problem
}

// NewProblemFactory returns a new ProblemFactory implementation.
func NewProblemFactory(config ProblemFactoryConfig) ProblemFactory {
	f := problemFactory{
		matchers:             config.Matchers,
		problemFactory:       config.ProblemFactory,
		statusProblemFactory: config.StatusProblemFactory,
		fallbackProblem:      config.FallbackProblem,
	}

	if f.problemFactory == nil {
		f.problemFactory = defaultProblemFactory{}
	}

	if f.statusProblemFactory == nil {
		if spf, ok := f.problemFactory.(StatusProblemFactory); ok {
			f.statusProblemFactory = spf
		} else {
			f.statusProblemFactory = defaultProblemFactory{}
		}
	}

	// Fallback problem intentionally has no default
	// A new problem is created each time, passing the context to the factory
	// A factory can attach correlation/request ID to every problem this way

	return f
}

// NewDefaultProblemFactory returns a new ProblemFactory implementation with default configuration.
func NewDefaultProblemFactory() ProblemFactory {
	return NewProblemFactory(ProblemFactoryConfig{})
}

func (f problemFactory) NewProblem(ctx context.Context, err error) problems.Problem {
	for _, matcher := range f.matchers {
		if matcher.MatchError(err) {
			if pf, ok := matcher.(ProblemFactory); ok {
				return pf.NewProblem(ctx, err)
			}

			if statusMatcher, ok := matcher.(StatusProblemMatcher); ok {
				if spf, ok := statusMatcher.(StatusProblemFactory); ok {
					return spf.NewStatusProblem(ctx, statusMatcher.Status(), err)
				}

				return f.statusProblemFactory.NewStatusProblem(ctx, statusMatcher.Status(), err)
			}

			return f.problemFactory.NewProblem(ctx, err)
		}
	}

	if f.fallbackProblem != nil {
		return f.fallbackProblem
	}

	return f.statusProblemFactory.NewStatusProblem(
		ctx,
		http.StatusInternalServerError,
		errors.New("something went wrong"),
	)
}
