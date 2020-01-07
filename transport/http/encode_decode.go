package http

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"

	"emperror.dev/errors"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/moogar0880/problems"
)

// NopResponseEncoder can be used for operations without output parameters.
// It returns 200 OK status code without a response body.
func NopResponseEncoder(_ context.Context, _ http.ResponseWriter, _ interface{}) error {
	return nil
}

// StatusCodeResponseEncoder can be used for operations without output parameters.
// It returns 200 OK status code without a response body.
func StatusCodeResponseEncoder(code int) kithttp.EncodeResponseFunc {
	return func(_ context.Context, w http.ResponseWriter, _ interface{}) error {
		w.WriteHeader(code)

		return nil
	}
}

// JSONResponseEncoder encodes the passed response object to the HTTP response writer in JSON format.
func JSONResponseEncoder(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	err := kithttp.EncodeJSONResponse(ctx, w, resp)
	if err != nil {
		return errors.Wrap(err, "failed to encode response")
	}

	return nil
}

// WithStatusCode wraps a response and implements the kithttp.StatusCoder interface.
// It allows passing a status code to kithttp.EncodeJSONResponse.
//
// Note: it only works with JSON marshaler.
func WithStatusCode(resp interface{}, code int) interface{} {
	return statusCodeResponseWrapper{
		response:   resp,
		statusCode: code,
	}
}

type statusCodeResponseWrapper struct {
	response   interface{}
	statusCode int
}

func (s statusCodeResponseWrapper) StatusCode() int {
	return s.statusCode
}

func (s statusCodeResponseWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.response)
}

// EncodeErrorResponseFunc encodes the passed error to the HTTP response writer.
// It's designed to be used in HTTP servers, for server-side endpoints.
// An EncodeErrorResponseFunc is supposed to return an error with the proper HTTP status code.
type EncodeErrorResponseFunc func(context.Context, http.ResponseWriter, error) error

// ErrorResponseEncoder encodes the passed response object to the HTTP response writer in JSON format.
func ErrorResponseEncoder(
	encoder kithttp.EncodeResponseFunc,
	errorEncoder EncodeErrorResponseFunc,
) kithttp.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
		if f, ok := resp.(endpoint.Failer); ok && f.Failed() != nil {
			return errorEncoder(ctx, w, f.Failed())
		}

		return encoder(ctx, w, resp)
	}
}

func errorResponseEncoderWrapper(errorEncoder EncodeErrorResponseFunc) kithttp.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		_ = errorEncoder(ctx, w, err)
	}
}

// ProblemConverter creates a new RFC-7807 Problem from an error.
type ProblemConverter interface {
	// NewProblem creates a new RFC-7807 Problem from an error.
	NewProblem(ctx context.Context, err error) problems.Problem
}

type defaultErrorProblemConverter struct{}

func (d defaultErrorProblemConverter) NewProblem(_ context.Context, _ error) problems.Problem {
	return problems.NewDetailedProblem(http.StatusInternalServerError, "something went wrong")
}

// NewJSONProblemErrorResponseEncoder returns an error response encoder that encodes errors following the
// RFC-7807 (Problem Details) standard (in JSON format).
//
// See details at https://tools.ietf.org/html/rfc7807
func NewJSONProblemErrorResponseEncoder(problemConverter ProblemConverter) EncodeErrorResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, err error) error {
		problem := problemConverter.NewProblem(ctx, err)

		w.Header().Set("Content-Type", problems.ProblemMediaType)
		if s, ok := problem.(problems.StatusProblem); ok && s.ProblemStatus() != 0 {
			w.WriteHeader(s.ProblemStatus())
		}

		return errors.WithStack(json.NewEncoder(w).Encode(problem))
	}
}

// NewDefaultJSONProblemErrorResponseEncoder returns an error response encoder that encodes errors following the
// RFC-7807 (Problem Details) standard (in JSON format).
//
// See details at https://tools.ietf.org/html/rfc7807
//
// The returned encoder encodes every error as 500 Internal Server Error.
func NewDefaultJSONProblemErrorResponseEncoder() EncodeErrorResponseFunc {
	return NewJSONProblemErrorResponseEncoder(defaultErrorProblemConverter{})
}

// NewXMLProblemErrorResponseEncoder returns an error response encoder that encodes errors following the
// RFC-7807 (Problem Details) standard (in XML format).
//
// See details at https://tools.ietf.org/html/rfc7807
func NewXMLProblemErrorResponseEncoder(problemConverter ProblemConverter) EncodeErrorResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, err error) error {
		problem := problemConverter.NewProblem(ctx, err)

		w.Header().Set("Content-Type", problems.ProblemMediaTypeXML)
		if s, ok := problem.(problems.StatusProblem); ok && s.ProblemStatus() != 0 {
			w.WriteHeader(s.ProblemStatus())
		}

		return errors.WithStack(xml.NewEncoder(w).Encode(problem))
	}
}

// NewDefaultXMLProblemErrorResponseEncoder returns an error response encoder that encodes errors following the
// RFC-7807 (Problem Details) standard (in XML format).
//
// See details at https://tools.ietf.org/html/rfc7807
//
// The returned encoder encodes every error as 500 Internal Server Error.
func NewDefaultXMLProblemErrorResponseEncoder() EncodeErrorResponseFunc {
	return NewXMLProblemErrorResponseEncoder(defaultErrorProblemConverter{})
}

// NewJSONProblemErrorEncoder returns an error encoder that encodes errors following the
// RFC-7807 (Problem Details) standard (in JSON format).
//
// See details at https://tools.ietf.org/html/rfc7807
func NewJSONProblemErrorEncoder(problemConverter ProblemConverter) kithttp.ErrorEncoder {
	return errorResponseEncoderWrapper(NewJSONProblemErrorResponseEncoder(problemConverter))
}

// NewDefaultJSONProblemErrorEncoder returns an error encoder that encodes errors following the
// RFC-7807 (Problem Details) standard (in JSON format).
//
// See details at https://tools.ietf.org/html/rfc7807
//
// The returned encoder encodes every error as 500 Internal Server Error.
func NewDefaultJSONProblemErrorEncoder() kithttp.ErrorEncoder {
	return errorResponseEncoderWrapper(NewDefaultJSONProblemErrorResponseEncoder())
}

// NewXMLProblemErrorEncoder returns an error encoder that encodes errors following the
// RFC-7807 (Problem Details) standard (in XML format).
//
// See details at https://tools.ietf.org/html/rfc7807
func NewXMLProblemErrorEncoder(problemConverter ProblemConverter) kithttp.ErrorEncoder {
	return errorResponseEncoderWrapper(NewXMLProblemErrorResponseEncoder(problemConverter))
}

// NewDefaultXMLProblemErrorEncoder returns an error encoder that encodes errors following the
// RFC-7807 (Problem Details) standard (in XML format).
//
// See details at https://tools.ietf.org/html/rfc7807
//
// The returned encoder encodes every error as 500 Internal Server Error.
func NewDefaultXMLProblemErrorEncoder() kithttp.ErrorEncoder {
	return errorResponseEncoderWrapper(NewDefaultXMLProblemErrorResponseEncoder())
}

// nolint: gochecknoglobals
var defaultJSONProblemErrorEncoder = NewDefaultJSONProblemErrorEncoder()

// ProblemErrorEncoder encodes errors in the Problem RFC format.
// Deprecated: use NewJSONProblemErrorEncoder or NewDefaultJSONProblemErrorEncoder instead.
func ProblemErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	defaultJSONProblemErrorEncoder(ctx, err, w)
}
