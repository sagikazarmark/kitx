package grpc

import (
	"google.golang.org/grpc/status"
)

// IsGRPCError checks if an error is already encoded as a gRPC status.
func IsGRPCError(err error) bool {
	// nolint: errorlint
	_, ok := err.(interface {
		GRPCStatus() *status.Status
	})

	return ok
}
