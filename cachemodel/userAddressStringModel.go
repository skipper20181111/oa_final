package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserAddressStringModel = (*customUserAddressStringModel)(nil)

type (
	// UserAddressStringModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAddressStringModel.
	UserAddressStringModel interface {
		userAddressStringModel
	}

	customUserAddressStringModel struct {
		*defaultUserAddressStringModel
	}
)

// NewUserAddressStringModel returns a model for the database table.
func NewUserAddressStringModel(conn sqlx.SqlConn) UserAddressStringModel {
	return &customUserAddressStringModel{
		defaultUserAddressStringModel: newUserAddressStringModel(conn),
	}
}
