package worker

import (
	"bookadmin/constants"
	"bookadmin/global"
	"bookadmin/model"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SyncWorker 同步Worker，负责消费Stream消息并批量写入MySQL
type SyncWorker struct {
	workerID      string
	consumerGroup string
	batchSize     int           // 批量处理大小
	batchTimeout  time.Duration // 批量超时时间
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewSyncWorker 创建同步Worker实例
func NewSyncWorker(workerID string) *SyncWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &SyncWorker{
		workerID:      workerID,
		consumerGroup: "sync-group",
		batchSize:     100,             // 每批处理100条
		batchTimeout:  5 * time.Second, // 5秒超时
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start 启动Worker
func (w *SyncWorker) Start() {
	global.GVA_LOG.Info(fmt.Sprintf("同步Worker [%s] 启动", w.workerID))

	// 启动两个goroutine分别处理点赞和收藏
	go w.consumeLikeStream()
	go w.consumeFavoriteStream()
}

// Stop 停止Worker
func (w *SyncWorker) Stop() {
	global.GVA_LOG.Info(fmt.Sprintf("同步Worker [%s] 停止", w.workerID))
	w.cancel()
}

// consumeLikeStream 消费点赞Stream
func (w *SyncWorker) consumeLikeStream() {
	streamName := constants.StreamLikeActions

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			// 从Stream读取消息
			messages, err := w.readStream(streamName)
			if err != nil {
				global.GVA_LOG.Error("读取点赞Stream失败",
					zap.String("worker", w.workerID),
					zap.Error(err))
				time.Sleep(time.Second) // 错误后等待1秒再重试
				continue
			}

			if len(messages) == 0 {
				time.Sleep(100 * time.Millisecond) // 没有消息，短暂等待
				continue
			}

			// 处理消息
			w.processLikeMessages(streamName, messages)
		}
	}
}

// consumeFavoriteStream 消费收藏Stream
func (w *SyncWorker) consumeFavoriteStream() {
	streamName := constants.StreamFavoriteActions

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			messages, err := w.readStream(streamName)
			if err != nil {
				global.GVA_LOG.Error("读取收藏Stream失败",
					zap.String("worker", w.workerID),
					zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			if len(messages) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			w.processFavoriteMessages(streamName, messages)
		}
	}
}

// readStream 从Stream读取消息
func (w *SyncWorker) readStream(streamName string) ([]redis.XMessage, error) {
	ctx, cancel := context.WithTimeout(w.ctx, 2*time.Second)
	defer cancel()

	// 使用XREADGROUP读取消息
	streams, err := global.GVA_REDIS.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    w.consumerGroup,
		Consumer: w.workerID,
		Streams:  []string{streamName, ">"},
		Count:    int64(w.batchSize),
		Block:    time.Second, // 阻塞1秒
	}).Result()

	if err == redis.Nil {
		return []redis.XMessage{}, nil
	}
	if err != nil {
		return nil, err
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return []redis.XMessage{}, nil
	}

	return streams[0].Messages, nil
}

// processLikeMessages 批量处理点赞消息
func (w *SyncWorker) processLikeMessages(streamName string, messages []redis.XMessage) {
	var likes []model.BookLike
	var unlikes []struct{ UserID, BookID uint }
	messageIDs := make([]string, 0, len(messages))

	// 解析消息
	for _, msg := range messages {
		userID, _ := strconv.ParseUint(msg.Values["user_id"].(string), 10, 32)
		bookID, _ := strconv.ParseUint(msg.Values["book_id"].(string), 10, 32)
		action := msg.Values["action"].(string)

		if action == "like" {
			likes = append(likes, model.BookLike{
				UserID: uint(userID),
				BookID: uint(bookID),
			})
		} else {
			unlikes = append(unlikes, struct{ UserID, BookID uint }{
				UserID: uint(userID),
				BookID: uint(bookID),
			})
		}

		messageIDs = append(messageIDs, msg.ID)
	}

	// 批量写入MySQL（使用事务）
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 处理点赞
		for _, like := range likes {
			// 使用INSERT IGNORE避免重复
			if err := tx.Create(&like).Error; err != nil {
				// 忽略唯一索引冲突错误
				continue
			}
			// 更新books表统计
			tx.Model(&model.Book{}).
				Where("id = ?", like.BookID).
				UpdateColumn("like_count", gorm.Expr("like_count + ?", 1))
		}

		// 处理取消点赞
		for _, unlike := range unlikes {
			if err := tx.Where("user_id = ? AND book_id = ?", unlike.UserID, unlike.BookID).
				Delete(&model.BookLike{}).Error; err != nil {
				continue
			}
			// 更新books表统计
			tx.Model(&model.Book{}).
				Where("id = ?", unlike.BookID).
				UpdateColumn("like_count", gorm.Expr("like_count - ?", 1))
		}

		return nil
	})

	if err != nil {
		global.GVA_LOG.Error("批量处理点赞消息失败",
			zap.String("worker", w.workerID),
			zap.Int("count", len(messages)),
			zap.Error(err))
		return
	}

	// 确认消息已处理（XACK）
	w.ackMessages(streamName, messageIDs)

	global.GVA_LOG.Info("批量处理点赞消息成功",
		zap.String("worker", w.workerID),
		zap.Int("likes", len(likes)),
		zap.Int("unlikes", len(unlikes)))
}

// processFavoriteMessages 批量处理收藏消息
func (w *SyncWorker) processFavoriteMessages(streamName string, messages []redis.XMessage) {
	var favorites []model.BookFavorite
	var unfavorites []struct{ UserID, BookID uint }
	messageIDs := make([]string, 0, len(messages))

	// 解析消息
	for _, msg := range messages {
		userID, _ := strconv.ParseUint(msg.Values["user_id"].(string), 10, 32)
		bookID, _ := strconv.ParseUint(msg.Values["book_id"].(string), 10, 32)
		action := msg.Values["action"].(string)

		if action == "favorite" {
			favorites = append(favorites, model.BookFavorite{
				UserID: uint(userID),
				BookID: uint(bookID),
			})
		} else {
			unfavorites = append(unfavorites, struct{ UserID, BookID uint }{
				UserID: uint(userID),
				BookID: uint(bookID),
			})
		}

		messageIDs = append(messageIDs, msg.ID)
	}

	// 批量写入MySQL
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for _, favorite := range favorites {
			if err := tx.Create(&favorite).Error; err != nil {
				continue
			}
			tx.Model(&model.Book{}).
				Where("id = ?", favorite.BookID).
				UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1))
		}

		for _, unfavorite := range unfavorites {
			if err := tx.Where("user_id = ? AND book_id = ?", unfavorite.UserID, unfavorite.BookID).
				Delete(&model.BookFavorite{}).Error; err != nil {
				continue
			}
			tx.Model(&model.Book{}).
				Where("id = ?", unfavorite.BookID).
				UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1))
		}

		return nil
	})

	if err != nil {
		global.GVA_LOG.Error("批量处理收藏消息失败",
			zap.String("worker", w.workerID),
			zap.Int("count", len(messages)),
			zap.Error(err))
		return
	}

	// 确认消息
	w.ackMessages(streamName, messageIDs)

	global.GVA_LOG.Info("批量处理收藏消息成功",
		zap.String("worker", w.workerID),
		zap.Int("favorites", len(favorites)),
		zap.Int("unfavorites", len(unfavorites)))
}

// ackMessages 确认消息已处理
func (w *SyncWorker) ackMessages(streamName string, messageIDs []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := global.GVA_REDIS.XAck(ctx, streamName, w.consumerGroup, messageIDs...).Err()
	if err != nil {
		global.GVA_LOG.Warn("确认Stream消息失败",
			zap.String("worker", w.workerID),
			zap.String("stream", streamName),
			zap.Error(err))
	}
}

// WorkerPool Worker池管理
type WorkerPool struct {
	workers []*SyncWorker
	size    int
}

// NewWorkerPool 创建Worker池
func NewWorkerPool(size int) *WorkerPool {
	return &WorkerPool{
		workers: make([]*SyncWorker, 0, size),
		size:    size,
	}
}

// Start 启动所有Worker
func (p *WorkerPool) Start() {
	for i := 0; i < p.size; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		worker := NewSyncWorker(workerID)
		worker.Start()
		p.workers = append(p.workers, worker)
	}
	global.GVA_LOG.Info(fmt.Sprintf("Worker池启动完成，共 %d 个Worker", p.size))
}

// Stop 停止所有Worker
func (p *WorkerPool) Stop() {
	for _, worker := range p.workers {
		worker.Stop()
	}
	global.GVA_LOG.Info("Worker池已停止")
}
