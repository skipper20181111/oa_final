package cachemodel

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CashAccountModel = (*customCashAccountModel)(nil)

type (
	// CashAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCashAccountModel.
	CashAccountModel interface {
		cashAccountModel
	}

	customCashAccountModel struct {
		*defaultCashAccountModel
	}
)

// NewCashAccountModel returns a model for the database table.
func NewCashAccountModel(conn sqlx.SqlConn, c cache.CacheConf) CashAccountModel {
	return &customCashAccountModel{
		defaultCashAccountModel: newCashAccountModel(conn, c),
	}
}
