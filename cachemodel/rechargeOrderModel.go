package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ RechargeOrderModel = (*customRechargeOrderModel)(nil)

type (
	// RechargeOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRechargeOrderModel.
	RechargeOrderModel interface {
		rechargeOrderModel
	}

	customRechargeOrderModel struct {
		*defaultRechargeOrderModel
	}
)

// NewRechargeOrderModel returns a model for the database table.
func NewRechargeOrderModel(conn sqlx.SqlConn) RechargeOrderModel {
	return &customRechargeOrderModel{
		defaultRechargeOrderModel: newRechargeOrderModel(conn),
	}
}
