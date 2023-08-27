package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SfPriceModel = (*customSfPriceModel)(nil)

type (
	// SfPriceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSfPriceModel.
	SfPriceModel interface {
		sfPriceModel
	}

	customSfPriceModel struct {
		*defaultSfPriceModel
	}
)

// NewSfPriceModel returns a model for the database table.
func NewSfPriceModel(conn sqlx.SqlConn) SfPriceModel {
	return &customSfPriceModel{
		defaultSfPriceModel: newSfPriceModel(conn),
	}
}
