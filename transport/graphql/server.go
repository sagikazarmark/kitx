package graphql

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport"
	"github.com/go-kit/log"
)

// Handler which should be called from the GraphQL binding of the service
// implementation. The incoming request parameter, and returned response
// parameter, are both GraphQL types, not user-domain.
type Handler interface {
	ServeGraphQL(ctx context.Context, request interface{}) (context.Context, interface{}, error)
}

// Server wraps an endpoint and implements graphql.Handler.
type Server struct {
	e            endpoint.Endpoint
	dec          DecodeRequestFunc
	enc          EncodeResponseFunc
	errorHandler transport.ErrorHandler
}

// NewServer constructs a new server, which implements wraps the provided
// endpoint and implements the Handler interface. Consumers should write
// bindings that adapt the concrete GraphQL queries and methods in the schema.
// Request and response objects are from the caller business domain, not GraphQL input and response types.
func NewServer(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	opts ...ServerOption,
) *Server {
	s := &Server{
		e:            e,
		dec:          dec,
		enc:          enc,
		errorHandler: transport.NewLogErrorHandler(log.NewNopLogger()),
	}

	for _, opt := range opts {
		opt.apply(s)
	}

	return s
}

// ServerOption sets an optional parameter for servers.
type ServerOption interface {
	apply(s *Server)
}

type serverOptionFunc func(s *Server)

func (fn serverOptionFunc) apply(s *Server) {
	fn(s)
}

// ServerErrorHandler is used to handle non-terminal errors. By default, non-terminal errors
// are ignored.
func ServerErrorHandler(errorHandler transport.ErrorHandler) ServerOption {
	return serverOptionFunc(func(s *Server) { s.errorHandler = errorHandler })
}

// ServeGraphQL implements the Handler interface.
func (s Server) ServeGraphQL(ctx context.Context, req interface{}) (context.Context, interface{}, error) {
	var (
		err         error
		request     interface{}
		response    interface{}
		graphqlResp interface{}
	)

	request, err = s.dec(ctx, req)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		return ctx, nil, err
	}

	response, err = s.e(ctx, request)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		return ctx, nil, err
	}

	graphqlResp, err = s.enc(ctx, response)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		return ctx, nil, err
	}

	return ctx, graphqlResp, nil
}
