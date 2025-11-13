package initialize

import (
	"bookadmin/service"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var cronScheduler *cron.Cron

// InitCronJobs 初始化定时任务
func InitCronJobs() {
	cronScheduler = cron.New(cron.WithSeconds()) // 支持秒级定时任务

	borrowService := service.NewBorrowService()
	reservationService := &service.ReservationService{}
	blacklistService := &service.BlacklistService{}

	// 每小时检查逾期记录
	_, err := cronScheduler.AddFunc("0 0 * * * *", func() {
		zap.L().Info("开始检查逾期记录...")
		if err := borrowService.CheckOverdueRecords(); err != nil {
			zap.L().Error("检查逾期记录失败", zap.Error(err))
		} else {
			zap.L().Info("逾期记录检查完成")
		}
	})
	if err != nil {
		zap.L().Error("添加逾期检查任务失败", zap.Error(err))
	}

	// 每天早上8点发送到期提醒
	_, err = cronScheduler.AddFunc("0 0 8 * * *", func() {
		zap.L().Info("开始发送到期提醒...")
		if err := borrowService.SendDueReminders(); err != nil {
			zap.L().Error("发送到期提醒失败", zap.Error(err))
		} else {
			zap.L().Info("到期提醒发送完成")
		}
	})
	if err != nil {
		zap.L().Error("添加到期提醒任务失败", zap.Error(err))
	}

	// 每小时检查过期预约
	_, err = cronScheduler.AddFunc("0 30 * * * *", func() {
		zap.L().Info("开始检查过期预约...")
		if err := reservationService.CheckExpiredReservations(); err != nil {
			zap.L().Error("检查过期预约失败", zap.Error(err))
		} else {
			zap.L().Info("过期预约检查完成")
		}
	})
	if err != nil {
		zap.L().Error("添加过期预约检查任务失败", zap.Error(err))
	}

	// 每天凌晨2点检查并自动拉黑逾期严重者
	_, err = cronScheduler.AddFunc("0 0 2 * * *", func() {
		zap.L().Info("开始检查并自动拉黑逾期严重者...")
		if err := blacklistService.CheckAndAddOverdueBlacklist(); err != nil {
			zap.L().Error("自动拉黑失败", zap.Error(err))
		} else {
			zap.L().Info("自动拉黑检查完成")
		}
	})
	if err != nil {
		zap.L().Error("添加自动拉黑任务失败", zap.Error(err))
	}

	// 每天凌晨3点检查过期黑名单
	_, err = cronScheduler.AddFunc("0 0 3 * * *", func() {
		zap.L().Info("开始检查过期黑名单...")
		if err := blacklistService.CheckExpiredBlacklist(); err != nil {
			zap.L().Error("检查过期黑名单失败", zap.Error(err))
		} else {
			zap.L().Info("过期黑名单检查完成")
		}
	})
	if err != nil {
		zap.L().Error("添加过期黑名单检查任务失败", zap.Error(err))
	}

	// 启动调度器
	cronScheduler.Start()
	zap.L().Info("定时任务调度器已启动")
}

// StopCronJobs 停止定时任务
func StopCronJobs() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		zap.L().Info("定时任务调度器已停止")
	}
}
