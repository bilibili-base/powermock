package util

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/bilibili-base/powermock/pkg/util/logger"
)

var codeMapping = map[int]int{
	0:  http.StatusOK,
	1:  http.StatusInternalServerError,
	2:  http.StatusInternalServerError,
	3:  http.StatusBadRequest,
	4:  http.StatusRequestTimeout,
	5:  http.StatusNotFound,
	6:  http.StatusConflict,
	7:  http.StatusForbidden,
	8:  http.StatusTooManyRequests,
	9:  http.StatusPreconditionFailed,
	10: http.StatusConflict,
	11: http.StatusBadRequest,
	12: http.StatusNotImplemented,
	13: http.StatusInternalServerError,
	14: http.StatusServiceUnavailable,
	15: http.StatusInternalServerError,
	16: http.StatusUnauthorized,
}

// GetHTTPCodeFromError is used to get http code from error
func GetHTTPCodeFromError(err error) int {
	return codeMapping[int(status.Code(err))]
}

// GRPCLoggingMiddleware is grpc logging middleware
func GRPCLoggingMiddleware(logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		method, _ := grpc.Method(ctx)
		reply, err := handler(ctx, req)
		md, _ := metadata.FromIncomingContext(ctx)
		if err != nil {
			logger.LogError(map[string]interface{}{
				"method":   method,
				"code":     status.Code(err),
				"metadata": md,
				"error":    err,
			}, "request received")
			return
		}
		logger.LogInfo(map[string]interface{}{
			"method":   method,
			"metadata": md,
			"code":     0,
		}, "request received")
		return reply, err
	}
}
