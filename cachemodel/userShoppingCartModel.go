package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserShoppingCartModel = (*customUserShoppingCartModel)(nil)

type (
	// UserShoppingCartModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserShoppingCartModel.
	UserShoppingCartModel interface {
		userShoppingCartModel
	}

	customUserShoppingCartModel struct {
		*defaultUserShoppingCartModel
	}
)

// NewUserShoppingCartModel returns a model for the database table.
func NewUserShoppingCartModel(conn sqlx.SqlConn) UserShoppingCartModel {
	return &customUserShoppingCartModel{
		defaultUserShoppingCartModel: newUserShoppingCartModel(conn),
	}
}
