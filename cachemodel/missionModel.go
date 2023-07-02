package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ MissionModel = (*customMissionModel)(nil)

type (
	// MissionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMissionModel.
	MissionModel interface {
		missionModel
	}

	customMissionModel struct {
		*defaultMissionModel
	}
)

// NewMissionModel returns a model for the database table.
func NewMissionModel(conn sqlx.SqlConn) MissionModel {
	return &customMissionModel{
		defaultMissionModel: newMissionModel(conn),
	}
}
