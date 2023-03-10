package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ CashLogModel = (*customCashLogModel)(nil)

type (
	// CashLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCashLogModel.
	CashLogModel interface {
		cashLogModel
	}

	customCashLogModel struct {
		*defaultCashLogModel
	}
)

// NewCashLogModel returns a model for the database table.
func NewCashLogModel(conn sqlx.SqlConn) CashLogModel {
	return &customCashLogModel{
		defaultCashLogModel: newCashLogModel(conn),
	}
}
