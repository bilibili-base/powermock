package util

import (
	"context"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// StartMetricsServer is used to start metric server
func StartMetricsServer(ctx context.Context, cancelFunc context.CancelFunc, logger logger.Logger, address string, registerer prometheus.Registerer) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	StartServiceAsync(ctx, cancelFunc, logger, func() error {
		return http.Serve(listener, promhttp.InstrumentMetricHandler(
			registerer, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}),
		))
	}, func() error {
		return listener.Close()
	})
	return nil
}
