package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ PointLogModel = (*customPointLogModel)(nil)

type (
	// PointLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPointLogModel.
	PointLogModel interface {
		pointLogModel
	}

	customPointLogModel struct {
		*defaultPointLogModel
	}
)

// NewPointLogModel returns a model for the database table.
func NewPointLogModel(conn sqlx.SqlConn) PointLogModel {
	return &customPointLogModel{
		defaultPointLogModel: newPointLogModel(conn),
	}
}
