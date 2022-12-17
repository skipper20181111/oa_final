package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserOrderModel = (*customUserOrderModel)(nil)

type (
	// UserOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserOrderModel.
	UserOrderModel interface {
		userOrderModel
	}

	customUserOrderModel struct {
		*defaultUserOrderModel
	}
)

// NewUserOrderModel returns a model for the database table.
func NewUserOrderModel(conn sqlx.SqlConn) UserOrderModel {
	return &customUserOrderModel{
		defaultUserOrderModel: newUserOrderModel(conn),
	}
}
