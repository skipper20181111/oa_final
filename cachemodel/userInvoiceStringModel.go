package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserInvoiceStringModel = (*customUserInvoiceStringModel)(nil)

type (
	// UserInvoiceStringModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserInvoiceStringModel.
	UserInvoiceStringModel interface {
		userInvoiceStringModel
	}

	customUserInvoiceStringModel struct {
		*defaultUserInvoiceStringModel
	}
)

// NewUserInvoiceStringModel returns a model for the database table.
func NewUserInvoiceStringModel(conn sqlx.SqlConn) UserInvoiceStringModel {
	return &customUserInvoiceStringModel{
		defaultUserInvoiceStringModel: newUserInvoiceStringModel(conn),
	}
}
