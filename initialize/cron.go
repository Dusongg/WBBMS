package initialize

import (
	"bookadmin/global"
	"bookadmin/service"
	"context"
	"fmt"
	"time"

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
		runCronJob("check_overdue_records", borrowService.CheckOverdueRecords)
	})
	if err != nil {
		zap.L().Error("添加逾期检查任务失败", zap.Error(err))
	}

	// 每天早上8点发送到期提醒
	_, err = cronScheduler.AddFunc("0 0 8 * * *", func() {
		runCronJob("send_due_reminders", borrowService.SendDueReminders)
	})
	if err != nil {
		zap.L().Error("添加到期提醒任务失败", zap.Error(err))
	}

	// 每小时检查过期预约
	_, err = cronScheduler.AddFunc("0 30 * * * *", func() {
		runCronJob("check_expired_reservations", reservationService.CheckExpiredReservations)
	})
	if err != nil {
		zap.L().Error("添加过期预约检查任务失败", zap.Error(err))
	}

	// 每天凌晨2点检查并自动拉黑逾期严重者
	_, err = cronScheduler.AddFunc("0 0 2 * * *", func() {
		runCronJob("check_and_add_overdue_blacklist", blacklistService.CheckAndAddOverdueBlacklist)
	})
	if err != nil {
		zap.L().Error("添加自动拉黑任务失败", zap.Error(err))
	}

	// 每天凌晨3点检查过期黑名单
	_, err = cronScheduler.AddFunc("0 0 3 * * *", func() {
		runCronJob("check_expired_blacklist", blacklistService.CheckExpiredBlacklist)
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

func runCronJob(name string, fn func() error) {
	if !acquireCronLock(name) {
		zap.L().Info("跳过定时任务，锁未获取", zap.String("job", name))
		return
	}
	defer releaseCronLock(name)

	start := time.Now()
	zap.L().Info("开始执行定时任务", zap.String("job", name))
	if err := fn(); err != nil {
		zap.L().Error("定时任务执行失败", zap.String("job", name), zap.Error(err))
		return
	}
	zap.L().Info("定时任务执行完成", zap.String("job", name), zap.Duration("duration", time.Since(start)))
}

func acquireCronLock(name string) bool {
	if global.GVA_CONFIG == nil || !global.GVA_CONFIG.Cron.UseDistributedLock || global.GVA_REDIS == nil {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	lockKey := fmt.Sprintf("lock:cron:%s", name)
	locked, err := global.GVA_REDIS.SetNX(ctx, lockKey, "1", time.Duration(global.GVA_CONFIG.Cron.LockTTLSeconds)*time.Second).Result()
	if err != nil {
		zap.L().Warn("获取定时任务分布式锁失败", zap.String("job", name), zap.Error(err))
		return false
	}
	return locked
}

func releaseCronLock(name string) {
	if global.GVA_CONFIG == nil || !global.GVA_CONFIG.Cron.UseDistributedLock || global.GVA_REDIS == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	lockKey := fmt.Sprintf("lock:cron:%s", name)
	if err := global.GVA_REDIS.Del(ctx, lockKey).Err(); err != nil {
		zap.L().Warn("释放定时任务分布式锁失败", zap.String("job", name), zap.Error(err))
	}
}
