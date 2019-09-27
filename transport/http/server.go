package http

import (
	"github.com/go-kit/kit/transport/http"
)

// ServerOptions collects a list of ServerOptions into a single option.
// Useful to avoid variadic hells when passing lists of options around.
func ServerOptions(options []http.ServerOption) http.ServerOption {
	return func(server *http.Server) {
		for _, option := range options {
			option(server)
		}
	}
}
