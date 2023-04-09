package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TransactionInfoModel = (*customTransactionInfoModel)(nil)

type (
	// TransactionInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTransactionInfoModel.
	TransactionInfoModel interface {
		transactionInfoModel
	}

	customTransactionInfoModel struct {
		*defaultTransactionInfoModel
	}
)

// NewTransactionInfoModel returns a model for the database table.
func NewTransactionInfoModel(conn sqlx.SqlConn) TransactionInfoModel {
	return &customTransactionInfoModel{
		defaultTransactionInfoModel: newTransactionInfoModel(conn),
	}
}
