package main

import (
	"bookadmin/initialize"
	"bookadmin/router"
	"bookadmin/worker"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	initialize.Zap()

	// 初始化数据库
	if initialize.GormMysql() == nil {
		zap.L().Error("数据库连接失败")
		return
	}

	// 初始化表结构
	initialize.Gorm()

	// 初始化配置缓存
	initialize.InitConfigCache()

	// 初始化Redis
	var workerPool *worker.WorkerPool
	if initialize.Redis() == nil {
		zap.L().Warn("Redis连接失败，点赞/收藏功能将受限")
	} else {
		// 初始化Redis Stream消费者组
		initialize.InitRedisStreamGroups()

		// 启动异步Worker池（5个Worker）
		workerPool = worker.NewWorkerPool(5)
		workerPool.Start()
	}

	// 初始化默认数据
	initialize.InitData()

	// 初始化定时任务
	initialize.InitCronJobs()

	// 初始化路由
	Router := router.InitRouter()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		zap.L().Info("收到关闭信号，正在优雅关闭...")
		if workerPool != nil {
			workerPool.Stop()
		}
		initialize.StopCronJobs()
		os.Exit(0)
	}()

	// 启动服务器
	port := 8888
	zap.L().Info(fmt.Sprintf("服务器启动在端口: %d", port))
	if err := Router.Run(fmt.Sprintf(":%d", port)); err != nil {
		zap.L().Error("服务器启动失败", zap.Error(err))
	}
}
