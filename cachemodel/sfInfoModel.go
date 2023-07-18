package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SfInfoModel = (*customSfInfoModel)(nil)

type (
	// SfInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSfInfoModel.
	SfInfoModel interface {
		sfInfoModel
	}

	customSfInfoModel struct {
		*defaultSfInfoModel
	}
)

// NewSfInfoModel returns a model for the database table.
func NewSfInfoModel(conn sqlx.SqlConn) SfInfoModel {
	return &customSfInfoModel{
		defaultSfInfoModel: newSfInfoModel(conn),
	}
}
