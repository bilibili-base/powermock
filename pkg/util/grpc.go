// Copyright 2021 bilibili-base
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
