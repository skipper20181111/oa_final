// Code generated by goctl. DO NOT EDIT!

package cachemodel

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	rechargeOrderFieldNames          = builder.RawFieldNames(&RechargeOrder{})
	rechargeOrderRows                = strings.Join(rechargeOrderFieldNames, ",")
	rechargeOrderRowsExpectAutoSet   = strings.Join(stringx.Remove(rechargeOrderFieldNames, "`id`", "`create_time`", "`update_time`", "`create_at`", "`update_at`"), ",")
	rechargeOrderRowsWithPlaceHolder = strings.Join(stringx.Remove(rechargeOrderFieldNames, "`id`", "`create_time`", "`update_time`", "`create_at`", "`update_at`"), "=?,") + "=?"
)

type (
	rechargeOrderModel interface {
		Insert(ctx context.Context, data *RechargeOrder) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*RechargeOrder, error)
		FindOneByOrderSn(ctx context.Context, orderSn string) (*RechargeOrder, error)
		FindOneByOutTradeNo(ctx context.Context, outTradeNo string) (*RechargeOrder, error)
		Update(ctx context.Context, data *RechargeOrder) error
		Delete(ctx context.Context, id int64) error
		UpdateFinished(ctx context.Context, OutTradeNo, TransactionId string, PaymentTime time.Time) error
	}

	defaultRechargeOrderModel struct {
		conn  sqlx.SqlConn
		table string
	}

	RechargeOrder struct {
		Id              int64     `db:"id"`                // id
		Phone           string    `db:"phone"`             // 账户手机号
		OrderSn         string    `db:"order_sn"`          // 订单编号
		OutTradeNo      string    `db:"out_trade_no"`      // 微信交易编号
		TransactionId   string    `db:"transaction_id"`    // 微信支付编号
		CreateOrderTime time.Time `db:"create_order_time"` // 订单产生时间
		Rpid            int64     `db:"rpid"`              // 充值id
		Amount          int64     `db:"amount"`            // 充值金额
		GiftAmount      int64     `db:"gift_amount"`       // 赠送金额
		WexinPayAmount  int64     `db:"wexin_pay_amount"`  // 微信支付金额
		OrderStatus     int64     `db:"order_status"`      // 订单状态：0->待付款；1->已完成
		PaymentTime     time.Time `db:"payment_time"`      // 支付时间
		LogId           int64     `db:"log_id"`
	}
)

func newRechargeOrderModel(conn sqlx.SqlConn) *defaultRechargeOrderModel {
	return &defaultRechargeOrderModel{
		conn:  conn,
		table: "`recharge_order`",
	}
}

func (m *defaultRechargeOrderModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *defaultRechargeOrderModel) FindOne(ctx context.Context, id int64) (*RechargeOrder, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", rechargeOrderRows, m.table)
	var resp RechargeOrder
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultRechargeOrderModel) FindOneByOrderSn(ctx context.Context, orderSn string) (*RechargeOrder, error) {
	var resp RechargeOrder
	query := fmt.Sprintf("select %s from %s where `order_sn` = ? limit 1", rechargeOrderRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, orderSn)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultRechargeOrderModel) FindOneByOutTradeNo(ctx context.Context, outTradeNo string) (*RechargeOrder, error) {
	var resp RechargeOrder
	query := fmt.Sprintf("select %s from %s where `out_trade_no` = ? limit 1", rechargeOrderRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, outTradeNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultRechargeOrderModel) Insert(ctx context.Context, data *RechargeOrder) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, rechargeOrderRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.Phone, data.OrderSn, data.OutTradeNo, data.TransactionId, data.CreateOrderTime, data.Rpid, data.Amount, data.GiftAmount, data.WexinPayAmount, data.OrderStatus, data.PaymentTime, data.LogId)
	return ret, err
}

func (m *defaultRechargeOrderModel) Update(ctx context.Context, newData *RechargeOrder) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, rechargeOrderRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, newData.Phone, newData.OrderSn, newData.OutTradeNo, newData.TransactionId, newData.CreateOrderTime, newData.Rpid, newData.Amount, newData.GiftAmount, newData.WexinPayAmount, newData.OrderStatus, newData.PaymentTime, newData.LogId, newData.Id)
	return err
}

func (m *defaultRechargeOrderModel) UpdateFinished(ctx context.Context, OutTradeNo, TransactionId string, PaymentTime time.Time) error {
	query := fmt.Sprintf("update %s set `order_status`=?,`payment_time`=?,`transaction_id`=? where `out_trade_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, int64(1), PaymentTime, TransactionId, OutTradeNo)
	return err
}

func (m *defaultRechargeOrderModel) tableName() string {
	return m.table
}
