package util

import (
	"context"

	"github.com/storyicon/powermock/pkg/util/logger"
)

// StartServiceAsync is used to start service async
func StartServiceAsync(ctx context.Context, cancelFunc context.CancelFunc, logger logger.Logger, serveFn func() error, stopFn func() error) {
	if serveFn == nil {
		return
	}
	go func() {
		logger.LogInfo(nil, "starting service")
		go func() {
			if err := serveFn(); err != nil {
				logger.LogError(nil, "error serving service: %s", err)
			}
			if cancelFunc != nil {
				cancelFunc()
			}
		}()
		<-ctx.Done()
		logger.LogInfo(nil, "stopping service")
		if stopFn() != nil {
			logger.LogInfo(nil, "stopping service gracefully")
			if err := stopFn(); err != nil {
				logger.LogWarn(nil, "error occurred while stopping service: %s", err)
			}
		}
		logger.LogInfo(nil, "exiting service")
	}()
}
