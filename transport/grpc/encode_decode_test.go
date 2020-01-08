package grpc

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type statusConverterStub struct {
	status *status.Status
}

func (s statusConverterStub) NewStatus(_ context.Context, _ error) *status.Status {
	return s.status
}

func TestStatusErrorResponseEncoder(t *testing.T) {
	statusConverter := statusConverterStub{status.New(codes.NotFound, "error")}

	errorEncoder := NewStatusErrorResponseEncoder(statusConverter)

	err := errorEncoder(context.Background(), errors.New("error"))

	s := status.Convert(err)

	if want, have := codes.NotFound, s.Code(); want != have {
		t.Errorf("unexpected code\nexpected: %d\nactual:   %d", want, have)
	}

	if want, have := "error", s.Message(); want != have {
		t.Errorf("unexpected message\nexpected: %s\nactual:   %s", want, have)
	}
}
