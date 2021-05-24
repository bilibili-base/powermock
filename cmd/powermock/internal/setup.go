package internal

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/storyicon/powermock/pkg/apimanager"
	grpcmockserver "github.com/storyicon/powermock/pkg/mockserver/grpc"
	httpmockserver "github.com/storyicon/powermock/pkg/mockserver/http"
	"github.com/storyicon/powermock/pkg/pluginregistry"
	pluginsgrpc "github.com/storyicon/powermock/pkg/pluginregistry/grpc"
	pluginshttp "github.com/storyicon/powermock/pkg/pluginregistry/http"
	pluginssimple "github.com/storyicon/powermock/pkg/pluginregistry/simple"
	pluginredis "github.com/storyicon/powermock/pkg/pluginregistry/storage/redis"
	"github.com/storyicon/powermock/pkg/util/logger"
)

// Startup is used to start up application
func Startup(
	ctx context.Context, cancelFunc context.CancelFunc,
	cfg *Config, log logger.Logger, registerer prometheus.Registerer) error {

	log.LogInfo(nil, "* start to create pluginRegistry")
	pluginRegistry, err := pluginregistry.New(cfg.PluginRegistry, log, registerer)
	if err != nil {
		return err
	}

	log.LogInfo(nil, "* start to create apiManager")
	apiManager, err := apimanager.New(cfg.ApiManager,
		pluginRegistry,
		log, registerer)
	if err != nil {
		log.LogFatal(nil, "failed to creating apiManager: %s", err)
	}

	log.LogInfo(nil, "* start to create grpcMockServer")
	grpcMockServer, err := grpcmockserver.New(
		cfg.GRPCMockServer,
		apiManager,
		log,
		prometheus.DefaultRegisterer,
	)
	if err != nil {
		log.LogFatal(nil, "failed to creating gRPCMockServer:", err)
	}

	log.LogInfo(nil, "* start to create httpMockServer")
	httpMockServer, err := httpmockserver.New(
		cfg.HTTPMockServer,
		apiManager,
		log,
		prometheus.DefaultRegisterer,
	)
	if err != nil {
		log.LogFatal(nil, "failed to creating httpMockServer: %s", err)
	}

	log.LogInfo(nil, "* start to create plugin(redis)")
	storagePlugin, err := pluginredis.New(cfg.Plugin.Redis, log, registerer)
	if err != nil {
		return err
	}

	log.LogInfo(nil, "* start to create plugin(simple)")
	simplePlugin, err := pluginssimple.New(cfg.Plugin.Simple, log, registerer)
	if err != nil {
		return err
	}

	log.LogInfo(nil, "* start to create plugin(gRPC)")
	grpcPlugin, err := pluginsgrpc.New(cfg.Plugin.GRPC, grpcMockServer.GetProtoManager().GetMethod, log, registerer)
	if err != nil {
		return err
	}

	log.LogInfo(nil, "* start to create plugin(http)")
	httpPlugin, err := pluginshttp.New(cfg.Plugin.HTTP, log, registerer)
	if err != nil {
		return err
	}

	log.LogInfo(nil, "* start to install plugins")
	if err := pluginRegistry.RegisterMockPlugins(simplePlugin, grpcPlugin, httpPlugin); err != nil {
		return err
	}
	if err := pluginRegistry.RegisterMatchPlugins(simplePlugin); err != nil {
		return err
	}
	if err := pluginRegistry.RegisterStoragePlugins(storagePlugin); err != nil {
		return err
	}

	log.LogInfo(nil, "* start to start apiManager")
	if err := apiManager.Start(ctx, cancelFunc); err != nil {
		return err
	}

	log.LogInfo(nil, "* start to start gRPCMockServer")
	if err := grpcMockServer.Start(ctx, cancelFunc); err != nil {
		return err
	}

	log.LogInfo(nil, "* start to start httpMockServer")
	if err := httpMockServer.Start(ctx, cancelFunc); err != nil {
		return err
	}
	return nil
}
