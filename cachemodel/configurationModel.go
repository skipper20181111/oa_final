package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ConfigurationModel = (*customConfigurationModel)(nil)

type (
	// ConfigurationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customConfigurationModel.
	ConfigurationModel interface {
		configurationModel
	}

	customConfigurationModel struct {
		*defaultConfigurationModel
	}
)

// NewConfigurationModel returns a model for the database table.
func NewConfigurationModel(conn sqlx.SqlConn) ConfigurationModel {
	return &customConfigurationModel{
		defaultConfigurationModel: newConfigurationModel(conn),
	}
}
