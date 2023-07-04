package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ RefundInfoModel = (*customRefundInfoModel)(nil)

type (
	// RefundInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRefundInfoModel.
	RefundInfoModel interface {
		refundInfoModel
	}

	customRefundInfoModel struct {
		*defaultRefundInfoModel
	}
)

// NewRefundInfoModel returns a model for the database table.
func NewRefundInfoModel(conn sqlx.SqlConn) RefundInfoModel {
	return &customRefundInfoModel{
		defaultRefundInfoModel: newRefundInfoModel(conn),
	}
}
