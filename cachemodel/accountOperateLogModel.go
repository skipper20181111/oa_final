package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ AccountOperateLogModel = (*customAccountOperateLogModel)(nil)

type (
	// AccountOperateLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAccountOperateLogModel.
	AccountOperateLogModel interface {
		accountOperateLogModel
	}

	customAccountOperateLogModel struct {
		*defaultAccountOperateLogModel
	}
)

// NewAccountOperateLogModel returns a model for the database table.
func NewAccountOperateLogModel(conn sqlx.SqlConn) AccountOperateLogModel {
	return &customAccountOperateLogModel{
		defaultAccountOperateLogModel: newAccountOperateLogModel(conn),
	}
}
