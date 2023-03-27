package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ RechargeProductModel = (*customRechargeProductModel)(nil)

type (
	// RechargeProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRechargeProductModel.
	RechargeProductModel interface {
		rechargeProductModel
	}

	customRechargeProductModel struct {
		*defaultRechargeProductModel
	}
)

// NewRechargeProductModel returns a model for the database table.
func NewRechargeProductModel(conn sqlx.SqlConn) RechargeProductModel {
	return &customRechargeProductModel{
		defaultRechargeProductModel: newRechargeProductModel(conn),
	}
}
