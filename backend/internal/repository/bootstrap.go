package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// bootstrapLockKey 启动引导互斥锁的 advisory lock 常量键。
// 选用固定大整数：所有实例对同一数据库竞争同一把锁，互不影响业务 advisory key 命名空间。
const bootstrapLockKey int64 = 8901234567000000001

// bootstrapRetryInterval 未抢到锁时的轮询间隔。
const bootstrapRetryInterval = time.Second

// WithBootstrapLock 在跨实例排他的启动引导锁内执行 fn（建表、建/对齐序列、种子数据）。
//
// 解决“旧库缺少编号序列时多个实例同时启动”的竞争：
//   - 抢到锁的实例执行 DDL/对齐/种子；
//   - 未抢到的实例不会直接退出，而是进入稳定等待（可中断、可观测），
//     待持锁实例完成后再获锁，此时表与序列均已存在，fn 内的 IF NOT EXISTS/幂等逻辑直接沿用；
//   - 锁是 PostgreSQL session 级 advisory lock，绑定在独立连接上，函数返回（含 panic/错误）
//     后连接关闭即自动释放；进程崩溃由数据库检测连接断开自动释放，不会死锁。
//
// 仅当数据库本身不可达（无法查询锁状态）时才返回错误；单纯抢不到锁会一直等待。
func WithBootstrapLock(db *gorm.DB, logger *slog.Logger, fn func() error) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("bootstrap lock get underlying db: %w", err)
	}

	ctx := context.Background()
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap lock acquire connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	waited := false
	for {
		var locked bool
		// pg_try_advisory_lock 非阻塞：拿不到立即返回 false，由我们决定重试，
		// 避免 pg_advisory_lock 长时间阻塞占用语句超时语义。
		if err := conn.QueryRowContext(ctx,
			"SELECT pg_try_advisory_lock($1)", bootstrapLockKey).Scan(&locked); err != nil {
			return fmt.Errorf("bootstrap lock try acquire: %w", err)
		}
		if locked {
			if waited {
				logger.Info("bootstrap lock acquired after waiting for peer instance")
			}
			break
		}
		if !waited {
			logger.Info("another instance is bootstrapping the database, waiting instead of exiting")
			waited = true
		}
		time.Sleep(bootstrapRetryInterval)
	}

	// 持锁期间执行引导；无论 fn 成功与否都显式释放（连接关闭也会兜底释放）。
	runErr := fn()
	var unlocked bool
	if err := conn.QueryRowContext(ctx,
		"SELECT pg_advisory_unlock($1)", bootstrapLockKey).Scan(&unlocked); err != nil {
		logger.Error("bootstrap advisory lock release failed (will be auto-released on connection close)",
			"error", err.Error())
	}
	if !unlocked {
		logger.Warn("bootstrap advisory lock was not held at release time")
	}
	return runErr
}
