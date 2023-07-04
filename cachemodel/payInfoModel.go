package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ PayInfoModel = (*customPayInfoModel)(nil)

type (
	// PayInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPayInfoModel.
	PayInfoModel interface {
		payInfoModel
	}

	customPayInfoModel struct {
		*defaultPayInfoModel
	}
)

// NewPayInfoModel returns a model for the database table.
func NewPayInfoModel(conn sqlx.SqlConn) PayInfoModel {
	return &customPayInfoModel{
		defaultPayInfoModel: newPayInfoModel(conn),
	}
}
