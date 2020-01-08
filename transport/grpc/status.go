package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// StatusMatcher matches an error.
// It is an alias to the ErrorMatcher interface.
type StatusMatcher interface {
	ErrorMatcher
}

// StatusCodeMatcher matches an error and returns the appropriate code code for it.
type StatusCodeMatcher interface {
	ErrorMatcher

	// Code returns the gRPC code code.
	Code() codes.Code
}

type statusMatcher struct {
	code    codes.Code
	matcher ErrorMatcher
}

// NewStatusCodeMatcher returns a new StatusCodeMatcher.
func NewStatusCodeMatcher(code codes.Code, matcher ErrorMatcher) StatusCodeMatcher {
	return statusMatcher{
		code:    code,
		matcher: matcher,
	}
}

func (m statusMatcher) MatchError(err error) bool {
	return m.matcher.MatchError(err)
}

func (m statusMatcher) Code() codes.Code {
	return m.code
}

// StatusCodeConverter creates a new gRPC code with a code from an error.
type StatusCodeConverter interface {
	// NewStatusWithCode creates a new gRPC code with a code from an error.
	NewStatusWithCode(ctx context.Context, code codes.Code, err error) *status.Status
}

type defaultStatusConverter struct{}

func (c defaultStatusConverter) NewStatus(_ context.Context, err error) *status.Status {
	return status.New(codes.Internal, err.Error())
}

func (c defaultStatusConverter) NewStatusWithCode(_ context.Context, code codes.Code, err error) *status.Status {
	return status.New(code, err.Error())
}

type statusConverter struct {
	matchers []StatusMatcher

	statusConverter       StatusConverter
	statusStatusConverter StatusCodeConverter
}

// StatusConverterConfig configures the StatusConverter implementation.
type StatusConverterConfig struct {
	// Matchers are used to match errors and create gRPC statuses.
	// If no matchers match the error (or no matchers are configured) a fallback code is created/returned.
	//
	// If a matcher also implements StatusConverter it is used instead of the builtin StatusConverter
	// for creating the code.
	//
	// If a matchers also implements StatusCodeMatcher
	// the builtin StatusCodeConverter is used for creating the code.
	Matchers []StatusMatcher

	// Status converters used for converting errors to gRPC statuses.
	StatusConverter     StatusConverter
	StatusCodeConverter StatusCodeConverter
}

// NewStatusConverter returns a new StatusConverter implementation.
func NewStatusConverter(config StatusConverterConfig) StatusConverter {
	c := statusConverter{
		matchers:              config.Matchers,
		statusConverter:       config.StatusConverter,
		statusStatusConverter: config.StatusCodeConverter,
	}

	if c.statusConverter == nil {
		c.statusConverter = defaultStatusConverter{}
	}

	if c.statusStatusConverter == nil {
		if spc, ok := c.statusConverter.(StatusCodeConverter); ok {
			c.statusStatusConverter = spc
		} else {
			c.statusStatusConverter = defaultStatusConverter{}
		}
	}

	return c
}

// NewDefaultStatusConverter returns a new StatusConverter implementation with default configuration.
func NewDefaultStatusConverter() StatusConverter {
	return NewStatusConverter(StatusConverterConfig{})
}

func (c statusConverter) NewStatus(ctx context.Context, err error) *status.Status {
	for _, matcher := range c.matchers {
		if matcher.MatchError(err) {
			if converter, ok := matcher.(StatusConverter); ok {
				return converter.NewStatus(ctx, err)
			}

			if statusMatcher, ok := matcher.(StatusCodeMatcher); ok {
				return c.statusStatusConverter.NewStatusWithCode(ctx, statusMatcher.Code(), err)
			}

			return c.statusConverter.NewStatus(ctx, err)
		}
	}

	return c.statusStatusConverter.NewStatusWithCode(
		ctx,
		codes.Internal,
		errors.New("something went wrong"),
	)
}
