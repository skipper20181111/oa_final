package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ VoucherModel = (*customVoucherModel)(nil)

type (
	// VoucherModel is an interface to be customized, add more methods here,
	// and implement the added methods in customVoucherModel.
	VoucherModel interface {
		voucherModel
	}

	customVoucherModel struct {
		*defaultVoucherModel
	}
)

// NewVoucherModel returns a model for the database table.
func NewVoucherModel(conn sqlx.SqlConn) VoucherModel {
	return &customVoucherModel{
		defaultVoucherModel: newVoucherModel(conn),
	}
}
