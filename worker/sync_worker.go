package worker

import (
	"bookadmin/constants"
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/observability"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SyncWorker struct {
	workerID         string
	consumerGroup    string
	batchSize        int
	batchTimeout     time.Duration
	blockTimeout     time.Duration
	reclaimInterval  time.Duration
	minIdle          time.Duration
	maxRetry         int64
	deadLetterStream string
	ctx              context.Context
	cancel           context.CancelFunc
}

type userBookAction struct {
	UserID uint
	BookID uint
}

func NewSyncWorker(workerID string) *SyncWorker {
	cfg := global.GVA_CONFIG.Worker
	ctx, cancel := context.WithCancel(context.Background())

	return &SyncWorker{
		workerID:         workerID,
		consumerGroup:    cfg.ConsumerGroup,
		batchSize:        cfg.BatchSize,
		batchTimeout:     time.Duration(cfg.BatchTimeoutSeconds) * time.Second,
		blockTimeout:     time.Duration(cfg.BlockTimeoutSeconds) * time.Second,
		reclaimInterval:  time.Duration(cfg.ReclaimIntervalSeconds) * time.Second,
		minIdle:          time.Duration(cfg.MinIdleSeconds) * time.Second,
		maxRetry:         int64(cfg.MaxRetry),
		deadLetterStream: cfg.DeadLetterStream,
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (w *SyncWorker) Start() {
	global.GVA_LOG.Info("同步Worker启动", zap.String("worker", w.workerID))
	go w.consumeStream(constants.StreamLikeActions, w.processLikeMessages)
	go w.consumeStream(constants.StreamFavoriteActions, w.processFavoriteMessages)
	go w.reclaimLoop(constants.StreamLikeActions, w.processLikeMessages)
	go w.reclaimLoop(constants.StreamFavoriteActions, w.processFavoriteMessages)
}

func (w *SyncWorker) Stop() {
	global.GVA_LOG.Info("同步Worker停止", zap.String("worker", w.workerID))
	w.cancel()
}

func (w *SyncWorker) consumeStream(streamName string, processor func(string, []redis.XMessage) error) {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			messages, err := w.readStream(streamName)
			if err != nil {
				observability.IncWorkerFailure(streamName, "read")
				global.GVA_LOG.Error("读取Stream失败", zap.String("worker", w.workerID), zap.String("stream", streamName), zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
			if len(messages) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if err := processor(streamName, messages); err != nil {
				observability.IncWorkerFailure(streamName, "process")
				global.GVA_LOG.Error("处理Stream消息失败", zap.String("worker", w.workerID), zap.String("stream", streamName), zap.Int("count", len(messages)), zap.Error(err))
			}
		}
	}
}

func (w *SyncWorker) reclaimLoop(streamName string, processor func(string, []redis.XMessage) error) {
	ticker := time.NewTicker(w.reclaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.reclaimPending(streamName, processor)
		}
	}
}

func (w *SyncWorker) reclaimPending(streamName string, processor func(string, []redis.XMessage) error) {
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Second)
	defer cancel()

	pendingSummary, err := global.GVA_REDIS.XPending(ctx, streamName, w.consumerGroup).Result()
	if err == nil {
		observability.SetWorkerPending(streamName, pendingSummary.Count)
	}

	pendingMessages, err := global.GVA_REDIS.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamName,
		Group:  w.consumerGroup,
		Start:  "-",
		End:    "+",
		Count:  int64(w.batchSize),
		Idle:   w.minIdle,
	}).Result()
	if err != nil && err != redis.Nil {
		observability.IncWorkerFailure(streamName, "pending_scan")
		global.GVA_LOG.Warn("扫描待处理消息失败", zap.String("worker", w.workerID), zap.String("stream", streamName), zap.Error(err))
		return
	}

	for _, pending := range pendingMessages {
		if pending.RetryCount > w.maxRetry {
			if err := w.movePendingToDLQ(streamName, pending.ID, fmt.Sprintf("retry_count_exceeded:%d", pending.RetryCount)); err != nil {
				observability.IncWorkerFailure(streamName, "dead_letter")
				global.GVA_LOG.Warn("转移死信消息失败", zap.String("worker", w.workerID), zap.String("stream", streamName), zap.String("message_id", pending.ID), zap.Error(err))
			}
			continue
		}

		claimed, err := global.GVA_REDIS.XClaim(ctx, &redis.XClaimArgs{
			Stream:   streamName,
			Group:    w.consumerGroup,
			Consumer: w.workerID,
			MinIdle:  w.minIdle,
			Messages: []string{pending.ID},
		}).Result()
		if err != nil || len(claimed) == 0 {
			if err != nil && err != redis.Nil {
				observability.IncWorkerFailure(streamName, "claim")
				global.GVA_LOG.Warn("认领待处理消息失败", zap.String("worker", w.workerID), zap.String("stream", streamName), zap.String("message_id", pending.ID), zap.Error(err))
			}
			continue
		}

		if err := processor(streamName, claimed); err != nil {
			observability.IncWorkerFailure(streamName, "reprocess")
			global.GVA_LOG.Warn("重试处理待处理消息失败", zap.String("worker", w.workerID), zap.String("stream", streamName), zap.String("message_id", pending.ID), zap.Error(err))
		}
	}
}

func (w *SyncWorker) readStream(streamName string) ([]redis.XMessage, error) {
	ctx, cancel := context.WithTimeout(w.ctx, w.blockTimeout+time.Second)
	defer cancel()

	streams, err := global.GVA_REDIS.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    w.consumerGroup,
		Consumer: w.workerID,
		Streams:  []string{streamName, ">"},
		Count:    int64(w.batchSize),
		Block:    w.blockTimeout,
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

func (w *SyncWorker) processLikeMessages(streamName string, messages []redis.XMessage) error {
	var likes []model.BookLike
	var unlikes []userBookAction
	messageIDs := make([]string, 0, len(messages))

	for _, msg := range messages {
		action, userID, bookID, err := parseMessage(msg)
		if err != nil {
			if dlqErr := w.moveMessageToDLQ(streamName, msg, err.Error()); dlqErr != nil {
				return dlqErr
			}
			continue
		}

		switch action {
		case "like":
			likes = append(likes, model.BookLike{UserID: userID, BookID: bookID})
		case "unlike":
			unlikes = append(unlikes, userBookAction{UserID: userID, BookID: bookID})
		default:
			if dlqErr := w.moveMessageToDLQ(streamName, msg, "unsupported_action:"+action); dlqErr != nil {
				return dlqErr
			}
			continue
		}

		messageIDs = append(messageIDs, msg.ID)
	}

	if len(messageIDs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(w.ctx, w.batchTimeout)
	defer cancel()

	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, like := range likes {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&like)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				continue
			}
			if err := tx.Model(&model.Book{}).Where("id = ?", like.BookID).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
				return err
			}
		}

		for _, unlike := range unlikes {
			result := tx.Where("user_id = ? AND book_id = ?", unlike.UserID, unlike.BookID).Delete(&model.BookLike{})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				continue
			}
			if err := tx.Model(&model.Book{}).Where("id = ?", unlike.BookID).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	w.ackMessages(streamName, messageIDs)
	observability.AddWorkerProcessed(streamName, "success", len(messageIDs))
	global.GVA_LOG.Info("批量处理点赞消息成功", zap.String("worker", w.workerID), zap.Int("likes", len(likes)), zap.Int("unlikes", len(unlikes)))
	return nil
}

func (w *SyncWorker) processFavoriteMessages(streamName string, messages []redis.XMessage) error {
	var favorites []model.BookFavorite
	var unfavorites []userBookAction
	messageIDs := make([]string, 0, len(messages))

	for _, msg := range messages {
		action, userID, bookID, err := parseMessage(msg)
		if err != nil {
			if dlqErr := w.moveMessageToDLQ(streamName, msg, err.Error()); dlqErr != nil {
				return dlqErr
			}
			continue
		}

		switch action {
		case "favorite":
			favorites = append(favorites, model.BookFavorite{UserID: userID, BookID: bookID})
		case "unfavorite":
			unfavorites = append(unfavorites, userBookAction{UserID: userID, BookID: bookID})
		default:
			if dlqErr := w.moveMessageToDLQ(streamName, msg, "unsupported_action:"+action); dlqErr != nil {
				return dlqErr
			}
			continue
		}

		messageIDs = append(messageIDs, msg.ID)
	}

	if len(messageIDs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(w.ctx, w.batchTimeout)
	defer cancel()

	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, favorite := range favorites {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&favorite)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				continue
			}
			if err := tx.Model(&model.Book{}).Where("id = ?", favorite.BookID).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error; err != nil {
				return err
			}
		}

		for _, unfavorite := range unfavorites {
			result := tx.Where("user_id = ? AND book_id = ?", unfavorite.UserID, unfavorite.BookID).Delete(&model.BookFavorite{})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				continue
			}
			if err := tx.Model(&model.Book{}).Where("id = ?", unfavorite.BookID).UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	w.ackMessages(streamName, messageIDs)
	observability.AddWorkerProcessed(streamName, "success", len(messageIDs))
	global.GVA_LOG.Info("批量处理收藏消息成功", zap.String("worker", w.workerID), zap.Int("favorites", len(favorites)), zap.Int("unfavorites", len(unfavorites)))
	return nil
}

func (w *SyncWorker) movePendingToDLQ(streamName, messageID, reason string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := global.GVA_REDIS.XRangeN(ctx, streamName, messageID, messageID, 1).Result()
	if err != nil || len(result) == 0 {
		return err
	}
	return w.moveMessageToDLQ(streamName, result[0], reason)
}

func (w *SyncWorker) moveMessageToDLQ(streamName string, msg redis.XMessage, reason string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	values := map[string]interface{}{
		"source_stream": streamName,
		"original_id":   msg.ID,
		"worker_id":     w.workerID,
		"reason":        reason,
		"payload":       fmt.Sprintf("%v", msg.Values),
	}

	if err := global.GVA_REDIS.XAdd(ctx, &redis.XAddArgs{
		Stream: w.deadLetterStream,
		Values: values,
	}).Err(); err != nil {
		return err
	}

	observability.AddWorkerProcessed(streamName, "dead_letter", 1)
	w.ackMessages(streamName, []string{msg.ID})
	return nil
}

func (w *SyncWorker) ackMessages(streamName string, messageIDs []string) {
	if len(messageIDs) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := global.GVA_REDIS.XAck(ctx, streamName, w.consumerGroup, messageIDs...).Err(); err != nil {
		observability.IncWorkerFailure(streamName, "ack")
		global.GVA_LOG.Warn("确认Stream消息失败", zap.String("worker", w.workerID), zap.String("stream", streamName), zap.Error(err))
	}
}

func parseMessage(msg redis.XMessage) (string, uint, uint, error) {
	actionValue, ok := msg.Values["action"]
	if !ok {
		return "", 0, 0, fmt.Errorf("missing_action")
	}
	userValue, ok := msg.Values["user_id"]
	if !ok {
		return "", 0, 0, fmt.Errorf("missing_user_id")
	}
	bookValue, ok := msg.Values["book_id"]
	if !ok {
		return "", 0, 0, fmt.Errorf("missing_book_id")
	}

	userID, err := parseUintValue(userValue)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid_user_id:%w", err)
	}
	bookID, err := parseUintValue(bookValue)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid_book_id:%w", err)
	}

	action := fmt.Sprint(actionValue)
	return action, userID, bookID, nil
}

func parseUintValue(value interface{}) (uint, error) {
	switch v := value.(type) {
	case string:
		parsed, err := strconv.ParseUint(v, 10, 32)
		return uint(parsed), err
	case int64:
		return uint(v), nil
	case int:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	case uint:
		return v, nil
	default:
		parsed, err := strconv.ParseUint(fmt.Sprint(v), 10, 32)
		return uint(parsed), err
	}
}

type WorkerPool struct {
	workers []*SyncWorker
	size    int
}

func NewWorkerPool(size int) *WorkerPool {
	if size <= 0 {
		size = global.GVA_CONFIG.Worker.PoolSize
	}
	return &WorkerPool{
		workers: make([]*SyncWorker, 0, size),
		size:    size,
	}
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.size; i++ {
		workerID := fmt.Sprintf("%s-%d", global.GVA_CONFIG.Worker.ConsumerPrefix, i+1)
		worker := NewSyncWorker(workerID)
		worker.Start()
		p.workers = append(p.workers, worker)
	}
	global.GVA_LOG.Info("Worker池启动完成", zap.Int("size", p.size))
}

func (p *WorkerPool) Stop() {
	for _, worker := range p.workers {
		worker.Stop()
	}
	global.GVA_LOG.Info("Worker池已停止")
}
