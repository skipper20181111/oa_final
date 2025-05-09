// Code generated by goctl. DO NOT EDIT!

package cachemodel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	orderFieldNames          = builder.RawFieldNames(&Order{})
	orderRows                = strings.Join(orderFieldNames, ",")
	orderRowsExpectAutoSet   = strings.Join(stringx.Remove(orderFieldNames, "`id`", "`create_time`", "`update_time`", "`create_at`", "`update_at`"), ",")
	orderRowsWithPlaceHolder = strings.Join(stringx.Remove(orderFieldNames, "`id`", "`create_time`", "`update_time`", "`create_at`", "`update_at`"), "=?,") + "=?"
)

type (
	orderModel interface {
		Insert(ctx context.Context, data *Order) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*Order, error)
		FindOneByOrderSn(ctx context.Context, orderSn string) (*Order, error)
		FindOneByOutRefundNo(ctx context.Context, outRefundNo string) (*Order, error)
		Update(ctx context.Context, data *Order) error
		Delete(ctx context.Context, id int64) error
		FindInvoiceByPhone(ctx context.Context, phone string, pagenumber int, InvoiceStatus []int64) ([]*Order, error)
		UpdateStatusByOrderSn(ctx context.Context, status int64, orderSn string) error
		UpdateReceivedByOrderSn(ctx context.Context, orderSn string) error
		UpdateClosedByOrderSn(ctx context.Context, orderSn string) error
		PrepareAllGoods(ctx context.Context, MarketID int64) error
		UpdateStatusByDeliverySn(ctx context.Context, Status, OriStatus int64, SfSn string) error
		UpdateStatusByOutTradeSn(ctx context.Context, status int64, OutTradeNo string) error
		UpdateClosedByOutTradeSn(ctx context.Context, OutTradeNo string) error
		RefundCash(ctx context.Context, orderSn string) error
		RefundWeChat(ctx context.Context, orderSn string) error
		FindAllByPhone(ctx context.Context, phone string, pagenumber int) ([]*Order, error)
		FindAllByOutTradeNo(ctx context.Context, OutTradeNo string) ([]*Order, error)
		UpdateWeChatPay(ctx context.Context, status int64, OutTradeNo string) error
		UpdateCashPay(ctx context.Context, status int64, OutTradeNo string) error
		FindCanChanged(ctx context.Context) ([]*Order, error)
		FindStatusBiggerThan1(ctx context.Context) ([]*Order, error)
		FindStatus2(ctx context.Context) ([]*Order, error)
		UpdateDeliver(ctx context.Context, DeliverSn, DeliverCompany, OrderSn string) error
		UpdateInvoice(ctx context.Context, OutTradeSn string, InvoiceStatus int64) error
		FindAllByOutTradeNos(ctx context.Context, phone string, PayInfos []*PayInfo) ([]*Order, error)
		DeleteByOutTradeSn(ctx context.Context, OutTradeSn string) error
		FindAllPointsOrder(ctx context.Context, phone string) ([]*Order, error)
		FindAllByOutTradeNoNotDeleted(ctx context.Context, OutTradeNo string) ([]*Order, error)
		UpdateRefund(ctx context.Context, orderSn string) error
		FindAll1001(ctx context.Context) ([]*Order, error)
		FindAll1002(ctx context.Context) ([]*Order, error)
		FindAllByStatus(ctx context.Context, status int64) ([]*Order, error)
		FindDelivering(ctx context.Context) ([]*Order, error)
		FindStatus3(ctx context.Context) ([]string, error)
		FindAllStatusByOutTradeNo(ctx context.Context, OutTradeNo string) ([]int64, error)
		FindOneBySfSn(ctx context.Context, SfSn string) (*Order, error)
		UpdateAddress(ctx context.Context, OrderSn, AddressStr string) error
		FindAllPointsCouponOrder(ctx context.Context, phone string) ([]*Order, error)
		FindDeliveredOuTradeSn(ctx context.Context, start, end time.Time) ([]string, error)
		FindDeliveredOuTradeSnHistory(ctx context.Context) ([]string, error)
		UpdateWeChatDeliveredByOutTradeSn(ctx context.Context, OutTradeNo string) error
		DeleteInvalidOrder(ctx context.Context) error
	}

	defaultOrderModel struct {
		conn  sqlx.SqlConn
		table string
	}

	Order struct {
		Id                   int64     `db:"id"`                      // id
		Phone                string    `db:"phone"`                   // 账户手机号
		OrderSn              string    `db:"order_sn"`                // 订单编号
		OutTradeNo           string    `db:"out_trade_no"`            // 微信交易编号
		OutRefundNo          string    `db:"out_refund_no"`           // 微信退款编号
		CreateOrderTime      time.Time `db:"create_order_time"`       // 订单产生时间
		Pidlist              string    `db:"pidlist"`                 // 订单商品列表
		OrderType            int64     `db:"order_type"`              // 订单类型 0->starmall;1->正常;2->预售;3->打折;4->预售且打折;
		OriginalAmount       int64     `db:"original_amount"`         // 原始金额
		PromotionAmount      int64     `db:"promotion_amount"`        // 实际总金额金额
		CouponAmount         int64     `db:"coupon_amount"`           // 优惠券抵扣金额
		UsedCouponinfo       string    `db:"used_couponinfo"`         // 使用的优惠券信息 空值意味着没有使用优惠券
		ActualAmount         int64     `db:"actual_amount"`           // 实际支付金额
		WexinPayAmount       int64     `db:"wexin_pay_amount"`        // 微信支付金额
		CashAccountPayAmount int64     `db:"cash_account_pay_amount"` // 现金账户支付金额
		FreightAmount        int64     `db:"freight_amount"`          // 运费金额
		Address              string    `db:"address"`                 // 收货人信息
		OrderNote            string    `db:"order_note"`              // 订单备注
		FinishWeixinpay      int64     `db:"finish_weixinpay"`        // 是否完成微信支付
		FinishAccountpay     int64     `db:"finish_accountpay"`       // 是否完成账户支付
		PointsOrder          int64     `db:"points_order"`            // 是否为积分兑换订单
		PointsAmount         int64     `db:"points_amount"`           // 使用积分额度
		OrderStatus          int64     `db:"order_status"`            // 订单状态：0->待付款；1->已付款；2->已发货；3->已完成；4->已关闭；5->无效订单；6->已退货未退钱；7->已退货已退钱;8->超时;9->已删除; 99->待复核,1001->仓库备货中不可退款;
		WexinDeliveryStatus  int64     `db:"wexin_delivery_status"`   // 0->未发货，(1) 待发货；(2) 已发货；(3) 确认收货；(4) 交易完成；(5) 已退款。
		WexinDeliveryTime    time.Time `db:"wexin_delivery_time"`     // 微信支付时间
		DeliveryCompany      string    `db:"delivery_company"`        // 物流公司(配送方式)
		DeliverySn           string    `db:"delivery_sn"`             // 物流单号
		AutoConfirmDay       int64     `db:"auto_confirm_day"`        // 自动确认时间（天）
		Growth               int64     `db:"growth"`                  // 可以活动的成长值，等于消费额
		InvoiceStatus        int64     `db:"invoice_status"`          // 处理：0->待付款；1->已填信息预开票状态；2->开票中；3->开票完成；4->开票失败
		ConfirmStatus        int64     `db:"confirm_status"`          // 确认收货状态：0->未确认；1->已确认
		DeleteStatus         int64     `db:"delete_status"`           // 删除状态：0->未删除；1->已删除
		PaymentTime          time.Time `db:"payment_time"`            // 支付时间
		DeliveryTime         time.Time `db:"delivery_time"`           // 发货时间
		ReceiveTime          time.Time `db:"receive_time"`            // 确认收货时间
		CloseTime            time.Time `db:"close_time"`              // 订单关闭时间
		ModifyTime           time.Time `db:"modify_time"`             // 修改时间
		MarketPlayerId       int64     `db:"market_player_id"`        // 商户id
		LogId                int64     `db:"log_id"`
		ProductInfo          string    `db:"product_info"` // 订单商品信息，方便发货人辨认
	}
)

func newOrderModel(conn sqlx.SqlConn) *defaultOrderModel {
	return &defaultOrderModel{
		conn:  conn,
		table: "`order`",
	}
}

func (m *defaultOrderModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
func (m *defaultOrderModel) DeleteInvalidOrder(ctx context.Context) error {
	query := fmt.Sprintf("delete from %s where `order_status` in(8,9)", m.table)
	_, err := m.conn.ExecCtx(ctx, query)
	return err
}

func (m *defaultOrderModel) FindOne(ctx context.Context, id int64) (*Order, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", orderRows, m.table)
	var resp Order
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

func (m *defaultOrderModel) FindOneByOrderSn(ctx context.Context, orderSn string) (*Order, error) {
	var resp Order
	query := fmt.Sprintf("select %s from %s where `order_sn` = ? limit 1", orderRows, m.table)
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

func (m *defaultOrderModel) FindOneByOutRefundNo(ctx context.Context, outRefundNo string) (*Order, error) {
	var resp Order
	query := fmt.Sprintf("select %s from %s where `out_refund_no` = ? limit 1", orderRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, outRefundNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) Insert(ctx context.Context, data *Order) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, orderRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.Phone, data.OrderSn, data.OutTradeNo, data.OutRefundNo, data.CreateOrderTime, data.Pidlist, data.OrderType, data.OriginalAmount, data.PromotionAmount, data.CouponAmount, data.UsedCouponinfo, data.ActualAmount, data.WexinPayAmount, data.CashAccountPayAmount, data.FreightAmount, data.Address, data.OrderNote, data.FinishWeixinpay, data.FinishAccountpay, data.PointsOrder, data.PointsAmount, data.OrderStatus, data.WexinDeliveryStatus, data.WexinDeliveryTime, data.DeliveryCompany, data.DeliverySn, data.AutoConfirmDay, data.Growth, data.InvoiceStatus, data.ConfirmStatus, data.DeleteStatus, data.PaymentTime, data.DeliveryTime, data.ReceiveTime, data.CloseTime, data.ModifyTime, data.MarketPlayerId, data.LogId, data.ProductInfo)
	return ret, err
}

func (m *defaultOrderModel) Update(ctx context.Context, newData *Order) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, orderRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, newData.Phone, newData.OrderSn, newData.OutTradeNo, newData.OutRefundNo, newData.CreateOrderTime, newData.Pidlist, newData.OrderType, newData.OriginalAmount, newData.PromotionAmount, newData.CouponAmount, newData.UsedCouponinfo, newData.ActualAmount, newData.WexinPayAmount, newData.CashAccountPayAmount, newData.FreightAmount, newData.Address, newData.OrderNote, newData.FinishWeixinpay, newData.FinishAccountpay, newData.PointsOrder, newData.PointsAmount, newData.OrderStatus, newData.WexinDeliveryStatus, newData.WexinDeliveryTime, newData.DeliveryCompany, newData.DeliverySn, newData.AutoConfirmDay, newData.Growth, newData.InvoiceStatus, newData.ConfirmStatus, newData.DeleteStatus, newData.PaymentTime, newData.DeliveryTime, newData.ReceiveTime, newData.CloseTime, newData.ModifyTime, newData.MarketPlayerId, newData.LogId, newData.ProductInfo, newData.Id)
	return err
}

func (m *defaultOrderModel) DeleteByOutTradeSn(ctx context.Context, OutTradeSn string) error {
	query := fmt.Sprintf("delete from %s where `out_trade_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, OutTradeSn)
	return err
}
func (m *defaultOrderModel) FindInvoiceByPhone(ctx context.Context, phone string, pagenumber int, InvoiceStatus []int64) ([]*Order, error) {
	if pagenumber <= 0 || pagenumber > 10 {
		return make([]*Order, 0), nil
	}
	InvoiceStatusStr := ""
	for _, status := range InvoiceStatus {
		InvoiceStatusStr = fmt.Sprintf("%s,%d", InvoiceStatusStr, status)
	}
	sheetlen := 10
	pagenumber = 1
	offset := sheetlen * (pagenumber - 1)
	query := fmt.Sprintf("select %s from %s where `phone` = ? and `order_status` in(3,4) and `invoice_status` in (%s) order by create_order_time desc  limit ? OFFSET ?", orderRows, m.table, InvoiceStatusStr[1:])
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, phone, sheetlen, offset)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindAllByPhone(ctx context.Context, phone string, pagenumber int) ([]*Order, error) {
	if pagenumber <= 0 || pagenumber > 10 {
		pagenumber = 1
	}
	sheetlen := 100
	pagenumber = 1
	offset := sheetlen * (pagenumber - 1)
	query := fmt.Sprintf("select %s from %s where `phone` = ? and `order_status`<99 and `order_status`<>8 and `order_status`<>9 order by create_order_time desc  limit ? OFFSET ?", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, phone, sheetlen, offset)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindAllPointsCouponOrder(ctx context.Context, phone string) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `phone` = ? and `points_order` in(1,2)  order by `create_order_time` desc limit ?", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, phone, 10)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindAllPointsOrder(ctx context.Context, phone string) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `phone` = ? and `points_order`=1  order by `create_order_time` desc limit ?", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, phone, 10)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindAllByOutTradeNos(ctx context.Context, phone string, PayInfos []*PayInfo) ([]*Order, error) {
	OutTradeNoList := [5]string{"1", "1", "1", "1", "1"}
	for i, info := range PayInfos {
		OutTradeNoList[i] = info.OutTradeNo
	}
	query := fmt.Sprintf("select %s from %s where `phone` = ? and `out_trade_no` in (?,?,?,?,?)  and `points_order`=0 and `order_status`<>8 and `order_status`<>9  order by `create_order_time` desc", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, phone, OutTradeNoList[0], OutTradeNoList[1], OutTradeNoList[2], OutTradeNoList[3], OutTradeNoList[4])
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindAllByOutTradeNoNotDeleted(ctx context.Context, OutTradeNo string) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `out_trade_no` = ? and `order_status`<>8 and `order_status`<>9", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, OutTradeNo)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, errors.New("no")
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindAllByStatus(ctx context.Context, status int64) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `order_status` = ? ", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, errors.New("no")
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) FindAll1002(ctx context.Context) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `order_status` = 1002 ", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, errors.New("no")
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindAllStatusByOutTradeNo(ctx context.Context, OutTradeNo string) ([]int64, error) {
	query := fmt.Sprintf("select distinct `order_status` from %s where `out_trade_no` = ? ", m.table)
	var resp []int64
	err := m.conn.QueryRowsCtx(ctx, &resp, query, OutTradeNo)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, errors.New("no")
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) FindAllByOutTradeNo(ctx context.Context, OutTradeNo string) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `out_trade_no` = ? ", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, OutTradeNo)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, errors.New("no")
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) RefundWeChat(ctx context.Context, orderSn string) error {
	query := fmt.Sprintf("update %s set `finish_weixinpay` = -1 where `order_sn` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, orderSn)
	return err
}
func (m *defaultOrderModel) UpdateRefund(ctx context.Context, orderSn string) error {
	query := fmt.Sprintf("update %s set `order_status` = ? , `modify_time`=? where `order_sn` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, 6, time.Now(), orderSn)
	return err
}
func (m *defaultOrderModel) RefundCash(ctx context.Context, orderSn string) error {
	query := fmt.Sprintf("update %s set `finish_accountpay` = -1 where `order_sn` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, orderSn)
	return err
}

func (m *defaultOrderModel) UpdateStatusByDeliverySn(ctx context.Context, Status, OriStatus int64, SfSn string) error {
	query := fmt.Sprintf("update %s set `order_status`=? where `delivery_sn` = ? and `order_status`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, Status, SfSn, OriStatus)
	return err
}
func (m *defaultOrderModel) PrepareAllGoods(ctx context.Context, MarketID int64) error {
	query := fmt.Sprintf("update %s set `order_status`=1001,`delivery_time`=? where `order_status` in(1,1000) and `market_player_id`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, time.Now(), MarketID)
	return err
}
func (m *defaultOrderModel) UpdateStatusByOrderSn(ctx context.Context, status int64, orderSn string) error {
	query := fmt.Sprintf("update %s set `order_status`=? where `order_sn` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, orderSn)
	return err
}
func (m *defaultOrderModel) UpdateReceivedByOrderSn(ctx context.Context, orderSn string) error {
	query := fmt.Sprintf("update %s set `order_status`=3,`receive_time`=?,`close_time`=? where `order_sn` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, time.Now(), time.Now(), orderSn)
	return err
}
func (m *defaultOrderModel) UpdateClosedByOrderSn(ctx context.Context, orderSn string) error {
	query := fmt.Sprintf("update %s set `order_status`=4,`close_time`=? where `order_sn` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, time.Now(), orderSn)
	return err
}

func (m *defaultOrderModel) FindAll1001(ctx context.Context) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `order_status` = 1001", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) UpdateStatusByOutTradeSn(ctx context.Context, status int64, OutTradeNo string) error {
	query := fmt.Sprintf("update %s set `order_status`=? where `out_trade_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, OutTradeNo)
	return err
}

func (m *defaultOrderModel) UpdateRefundStatusByOutTradeSn(ctx context.Context, OutTradeNo string) error {
	query := fmt.Sprintf("update %s set `order_status`=6 where `out_trade_no` = ? and `order_status` in(1,1000) ", m.table)
	_, err := m.conn.ExecCtx(ctx, query, OutTradeNo)
	return err
}

func (m *defaultOrderModel) UpdateClosedByOutTradeSn(ctx context.Context, OutTradeNo string) error {
	query := fmt.Sprintf("update %s set `order_status`=4,`close_time`=? where `out_trade_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, time.Now(), OutTradeNo)
	return err
}

func (m *defaultOrderModel) UpdateWeChatPay(ctx context.Context, status int64, OutTradeNo string) error {
	query := fmt.Sprintf("update %s set `finish_weixinpay`=? where `out_trade_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, OutTradeNo)
	return err
}
func (m *defaultOrderModel) UpdateCashPay(ctx context.Context, status int64, OutTradeNo string) error {
	query := fmt.Sprintf("update %s set `finish_accountpay`=? where `out_trade_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, OutTradeNo)
	return err
}

func (m *defaultOrderModel) UpdateDeliver(ctx context.Context, DeliverSn, DeliverCompany, OrderSn string) error {
	//query := fmt.Sprintf("update %s set `delivery_sn`=?,`delivery_company`=?,`order_status`=1002 where `order_sn` = ?", m.table)
	query := fmt.Sprintf("update %s set `delivery_sn`=?,`delivery_company`=?,`order_status`=1000 where `order_sn` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, DeliverSn, DeliverCompany, OrderSn)
	return err
}

func (m *defaultOrderModel) UpdateInvoice(ctx context.Context, OutTradeSn string, InvoiceStatus int64) error {
	query := fmt.Sprintf("update %s set `invoice_status`=? where `out_trade_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, InvoiceStatus, OutTradeSn)
	return err
}

func (m *defaultOrderModel) UpdateAddress(ctx context.Context, OrderSn, AddressStr string) error {
	query := fmt.Sprintf("update %s set `address`=?,`delivery_sn`='' where `order_sn` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, AddressStr, OrderSn)
	return err
}

func (m *defaultOrderModel) FindDelivering(ctx context.Context) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `order_status` in(1002,1003)", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) FindStatus3(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf("select distinct `out_trade_no` from %s where `order_status` =3", m.table)
	var resp []string
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindStatus2(ctx context.Context) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `order_status` =2 limit 100", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindCanChanged(ctx context.Context) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `order_status` in(0,6) limit 100", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) FindStatusBiggerThan1(ctx context.Context) ([]*Order, error) {
	query := fmt.Sprintf("select %s from %s where `order_status` in(1,1000,1001) and `delivery_sn`='' limit 100", orderRows, m.table)
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) FindOneBySfSn(ctx context.Context, SfSn string) (*Order, error) {
	var resp Order
	query := fmt.Sprintf("select %s from %s where `delivery_sn` = ? limit 1", orderRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, SfSn)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) FindDeliveredOuTradeSnHistory(ctx context.Context) ([]string, error) {

	var resp []string
	query := fmt.Sprintf("select distinct `out_trade_no` from %s where  `order_status` in(1001,1002,1003,2,3,4) and `wexin_delivery_status`=0   and `wexin_pay_amount`>0 ", m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) FindDeliveredOuTradeSn(ctx context.Context, start, end time.Time) ([]string, error) {

	var resp []string
	//query := fmt.Sprintf("select distinct `out_trade_no` from %s where `order_status` in(2) and  `wexin_delivery_status`=0 and  `delivery_time`>=? and `delivery_time`<=? and `wexin_pay_amount`>0", m.table)
	query := fmt.Sprintf("select distinct `out_trade_no` from %s where `order_status` in(2) and  `wexin_delivery_status`=0 and `wexin_pay_amount`>0", m.table)

	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *defaultOrderModel) UpdateWeChatDeliveredByOutTradeSn(ctx context.Context, OutTradeNo string) error {
	query := fmt.Sprintf("update %s set `wexin_delivery_status`=2,`wexin_delivery_time`=? where `out_trade_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, time.Now(), OutTradeNo)
	return err
}
func (m *defaultOrderModel) tableName() string {
	return m.table
}
