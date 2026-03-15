package main

import (
	appconfig "bookadmin/config"
	"bookadmin/global"
	"bookadmin/initialize"
	"bookadmin/observability"
	"bookadmin/resilience"
	"bookadmin/router"
	"bookadmin/worker"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	mode := flag.String("mode", "all", "运行模式: api|worker|scheduler|all|migrate")
	flag.Parse()

	if _, err := initialize.Config(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	resilience.InitBreakers()

	logger := initialize.Zap()
	defer func() {
		_ = logger.Sync()
	}()

	zap.L().Info("应用启动", zap.String("mode", *mode))

	// 分布式追踪：api/all 模式需在 GORM 初始化前启用，以便 otelgorm 插件生效
	if *mode == "api" || *mode == "all" {
		if err := observability.InitTracer(context.Background()); err != nil {
			zap.L().Warn("分布式追踪初始化失败，将不发送链路数据", zap.Error(err))
		} else if observability.TracerEnabled() {
			zap.L().Info("分布式追踪已启用")
		}
	}

	// 初始化数据库
	if initialize.GormMysql() == nil {
		zap.L().Error("数据库连接失败")
		os.Exit(1)
	}

	if mustRunMigrations(*mode) {
		initialize.Gorm()
	}

	if shouldSeedData(*mode) {
		initialize.InitData()
	}

	if *mode == "migrate" {
		zap.L().Info("数据库迁移完成")
		return
	}

	initialize.InitConfigCache()

	redisReady := false
	var workerPool *worker.WorkerPool
	if initialize.Redis() == nil {
		zap.L().Warn("Redis连接失败，点赞/收藏功能将受限")
	} else {
		redisReady = true
		initialize.InitRedisStreamGroups()
	}

	if requiresRedis(*mode) && !redisReady {
		zap.L().Error("当前模式依赖Redis，但Redis不可用", zap.String("mode", *mode))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if observability.TracerEnabled() {
		defer func() {
			if err := observability.ShutdownTracer(context.Background()); err != nil {
				zap.L().Warn("Tracer关闭异常", zap.Error(err))
			}
		}()
	}

	var server *http.Server
	serverErrors := make(chan error, 1)

	switch *mode {
	case "api":
		server = startAPIServer(serverErrors)
	case "worker":
		workerPool = startWorkerPool()
	case "scheduler":
		startScheduler()
	case "all":
		server = startAPIServer(serverErrors)
		workerPool = startWorkerPool()
		startScheduler()
	default:
		zap.L().Error("未知运行模式", zap.String("mode", *mode))
		os.Exit(1)
	}

	select {
	case <-ctx.Done():
		zap.L().Info("收到关闭信号，开始优雅停机")
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Error("服务器启动失败", zap.Error(err))
		}
	}

	if workerPool != nil {
		workerPool.Stop()
	}
	initialize.StopCronJobs()

	if server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig().Server.ShutdownTimeoutSeconds)*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			zap.L().Error("HTTP服务优雅停机失败", zap.Error(err))
		}
	}
}

func mustRunMigrations(mode string) bool {
	return mode == "migrate" || appConfig().Migration.AutoMigrate
}

func shouldSeedData(mode string) bool {
	if !appConfig().Security.SeedDefaultUsers {
		return false
	}
	return mode == "migrate" || mode == "all" || mode == "api"
}

func requiresRedis(mode string) bool {
	return mode == "worker"
}

func startAPIServer(serverErrors chan<- error) *http.Server {
	cfg := appConfig()
	httpServer := &http.Server{
		Addr:         cfg.ServerAddress(),
		Handler:      router.InitRouter(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
	}

	go func() {
		zap.L().Info("HTTP服务启动", zap.String("addr", cfg.ServerAddress()))
		serverErrors <- httpServer.ListenAndServe()
	}()

	return httpServer
}

func startWorkerPool() *worker.WorkerPool {
	if !appConfig().Worker.Enabled {
		zap.L().Warn("Worker已在配置中禁用")
		return nil
	}

	workerPool := worker.NewWorkerPool(appConfig().Worker.PoolSize)
	workerPool.Start()
	return workerPool
}

func startScheduler() {
	if !appConfig().Cron.Enabled {
		zap.L().Warn("定时任务调度器已在配置中禁用")
		return
	}

	initialize.InitCronJobs()
}

func appConfig() *appconfig.Config {
	return global.GVA_CONFIG
}
