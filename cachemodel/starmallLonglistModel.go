package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ StarmallLonglistModel = (*customStarmallLonglistModel)(nil)

type (
	// StarmallLonglistModel is an interface to be customized, add more methods here,
	// and implement the added methods in customStarmallLonglistModel.
	StarmallLonglistModel interface {
		starmallLonglistModel
	}

	customStarmallLonglistModel struct {
		*defaultStarmallLonglistModel
	}
)

// NewStarmallLonglistModel returns a model for the database table.
func NewStarmallLonglistModel(conn sqlx.SqlConn) StarmallLonglistModel {
	return &customStarmallLonglistModel{
		defaultStarmallLonglistModel: newStarmallLonglistModel(conn),
	}
}
