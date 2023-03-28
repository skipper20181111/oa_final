package cachemodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ InvoiceModel = (*customInvoiceModel)(nil)

type (
	// InvoiceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customInvoiceModel.
	InvoiceModel interface {
		invoiceModel
	}

	customInvoiceModel struct {
		*defaultInvoiceModel
	}
)

// NewInvoiceModel returns a model for the database table.
func NewInvoiceModel(conn sqlx.SqlConn) InvoiceModel {
	return &customInvoiceModel{
		defaultInvoiceModel: newInvoiceModel(conn),
	}
}
