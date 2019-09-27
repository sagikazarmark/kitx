package grpc

import (
	"github.com/go-kit/kit/transport/grpc"
)

// ServerOptions collects a list of ServerOptions into a single option.
// Useful to avoid variadic hells when passing lists of options around.
func ServerOptions(options []grpc.ServerOption) grpc.ServerOption {
	return func(server *grpc.Server) {
		for _, option := range options {
			option(server)
		}
	}
}
