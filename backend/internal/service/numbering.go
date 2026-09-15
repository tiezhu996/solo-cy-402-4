package service

import (
	"fmt"
	"time"

	"cylawcase/internal/repository"
)

// 业务编号格式：前缀 + 年份 + 序列值。
// 唯一性由 PostgreSQL sequence（case_no_seq / bill_no_seq）保证，
// 序列跨多次启动、多个实例、同一数据库连续运行单调分配；时间戳仅承担可读性，不参与唯一性，
// 因此不同实例时钟不一致或重启都不会产生重号。
const (
	caseNoPrefix = "CY"
	billNoPrefix = "BILL"
)

// nextCaseNo 取案件编号。
func nextCaseNo(tx *repository.TxRepos) (string, error) {
	seq, err := tx.Sequences.NextValue(repository.CaseNoSeqName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%4d%08d", caseNoPrefix, time.Now().Year(), seq), nil
}

// nextBillNo 取账单编号。
func nextBillNo(tx *repository.TxRepos) (string, error) {
	seq, err := tx.Sequences.NextValue(repository.BillNoSeqName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%4d%010d", billNoPrefix, time.Now().Year(), seq), nil
}
