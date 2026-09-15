package service

import (
	"errors"
	"fmt"
	"log/slog"

	"cylawcase/internal/constants"
	"cylawcase/internal/repository"
	"cylawcase/internal/util"
)

// numberAllocateMaxAttempts 编号分配（含唯一冲突重试）的最大事务尝试次数。
// 每次尝试都使用全新 nextval，撞号概率本就为零；保留少量重试仅用于兜底手工历史脏号等极端情况。
const numberAllocateMaxAttempts = 4

// runWithNumberTx 包装“先取业务编号、再写库”的事务，统一处理编号冲突重试。
//   - fn 在单个事务内执行，应通过 tx.Sequences.NextValue 取号；
//   - 编号撞唯一索引（23505）时：整个事务回滚（本次 nextval 作废、绝不复用旧号），用新号重试；
//   - nextval 本身失败：属于可重试的临时错误，包装为 50301 返回，调用方稍后重试即可，且不落任何记录；
//   - 业务错误（*util.AppError，如案件已归档、客户不存在）：立即原样返回，绝不重试；
//   - 超过最大尝试次数仍冲突：返回 50301 可重试错误，保证失败语义稳定、成功时严格只写一条。
//
// 每次重试都是一个全新事务：对于建账场景会重新锁定案件行并复核“是否已归档”，
// 因此重试不会破坏已归档案件的费用终态。
func runWithNumberTx(logger *slog.Logger, txStore *repository.TxStore, entity string,
	fn func(tx *repository.TxRepos) error) error {
	var lastConflict error
	for attempt := 1; attempt <= numberAllocateMaxAttempts; attempt++ {
		err := txStore.WithinTx(fn)
		if err == nil {
			return nil
		}
		// 业务校验错误（含已归档禁止建账 40904）直接透传，不重试、不吞错。
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			return err
		}
		// 编号唯一冲突：回滚已由事务完成，换一个全新 nextval 重试。
		if repository.IsUniqueViolation(err) {
			lastConflict = err
			logger.Warn("business number conflict, retry with fresh sequence value",
				"entity", entity, "attempt", attempt, "error", err.Error())
			continue
		}
		// 取号失败等其他数据库错误：可重试的临时失败，不能落成 500 让调用方误以为数据已坏。
		return util.NewAppError(constants.CodeNumberAllocateFailed,
			fmt.Sprintf("%s number allocate failed (retryable): %s", entity, err.Error()))
	}
	// 多次冲突仍未成功：返回稳定的可重试错误，不落任何记录；下一次请求会用全新 nextval。
	return util.NewAppError(constants.CodeNumberAllocateFailed,
		fmt.Sprintf("%s number allocate conflict (retryable) after %d attempts: %s",
			entity, numberAllocateMaxAttempts, lastConflict))
}
