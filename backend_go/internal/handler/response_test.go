package handler

import (
	"net/http"
	"testing"

	"github.com/prejudicing/news_py_go/backend_go/internal/service"
)

// TestStatusForKeepsHTTPMappingAtTransportBoundary 验证业务错误类别只在 HTTP 层转换为状态码。
func TestStatusForKeepsHTTPMappingAtTransportBoundary(t *testing.T) {
	tests := []struct {
		kind     service.ErrorKind
		wantCode int
	}{
		{service.ErrorInvalidArgument, http.StatusUnprocessableEntity},
		{service.ErrorUnauthenticated, http.StatusUnauthorized},
		{service.ErrorNotFound, http.StatusNotFound},
		{service.ErrorUnavailable, http.StatusServiceUnavailable},
		{service.ErrorUpstream, http.StatusBadGateway},
		{service.ErrorInternal, http.StatusInternalServerError},
	}
	for _, test := range tests {
		if got := statusFor(test.kind); got != test.wantCode {
			t.Errorf("statusFor(%q) = %d, want %d", test.kind, got, test.wantCode)
		}
	}
}
