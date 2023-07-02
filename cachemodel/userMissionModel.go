package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserMissionModel = (*customUserMissionModel)(nil)

type (
	// UserMissionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserMissionModel.
	UserMissionModel interface {
		userMissionModel
	}

	customUserMissionModel struct {
		*defaultUserMissionModel
	}
)

// NewUserMissionModel returns a model for the database table.
func NewUserMissionModel(conn sqlx.SqlConn) UserMissionModel {
	return &customUserMissionModel{
		defaultUserMissionModel: newUserMissionModel(conn),
	}
}
