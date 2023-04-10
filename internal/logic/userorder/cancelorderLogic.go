package userorder

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"log"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelorderLogic {
	return &CancelorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelorderLogic) Cancelorder(req *types.CancelOrderRes) (resp *types.CancelOrderResp, err error) {
	lu := NewLogic(l.ctx, l.svcCtx)
	userphone := l.ctx.Value("phone").(string)
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	info, err := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if sn2order == nil || info == nil {
		return &types.CancelOrderResp{Code: "10000", Msg: "未查询到订单"}, nil
	}
	if sn2order.OrderStatus != 1 {
		return &types.CancelOrderResp{Code: "10000", Msg: "已发货无法退款"}, nil
	}
	sn2order.WexinPayAmount = info.WexinPayAmount
	sn2order.CashAccountPayAmount = info.CashAccountPayAmount

	lu.Oplog("微信退款", userphone, "开始更新", sn2order.LogId)
	service := refunddomestic.RefundsApiService{Client: l.svcCtx.Client}
	create, result, err := service.Create(l.ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(sn2order.OutTradeNo),
		OutRefundNo: core.String(sn2order.OutTradeNo),
		Amount: &refunddomestic.AmountReq{Currency: core.String("CNY"),
			Refund: core.Int64(sn2order.WexinPayAmount),
			Total:  core.Int64(sn2order.WexinPayAmount)},
	})
	defer result.Response.Body.Close()
	if err != nil {
		log.Printf("call Create err:%s", err)
		return &types.CancelOrderResp{Code: "4004", Msg: err.Error()}, nil
	} else {
		log.Printf("status=%d resp=%s", result.Response.StatusCode, resp, create.String())
	}
	lu.Oplog("微信退款", userphone, "结束更新", sn2order.LogId)
	//从这里开始更新现金账户于优惠券账户
	// 此时还有特别重要的事情，1，要更改现金账户余额，2，要更改优惠券账户，毕竟优惠券账户已经用完了。
	if sn2order.CashAccountPayAmount > 0 || sn2order.UsedCouponinfo != "" {
		lockmsglist := make([]*types.LockMsg, 0)
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.ctx.Value("phone").(string), Field: "user_coupon"})
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.ctx.Value("phone").(string), Field: "cash_account"})
		if lu.getlocktry(lockmsglist) {

			if sn2order.CashAccountPayAmount > 0 {
				ok, _ := lu.Updatecashaccount(sn2order, false)
				if !ok {
					lu.Oplog("更新现金账户失败", userphone+sn2order.OutTradeNo, "开始更新", sn2order.LogId)
				}
			}
			if sn2order.UsedCouponinfo != "" {
				ok, _ := lu.UpdateCoupon(sn2order, false)
				if !ok {
					lu.Oplog("更新优惠券失败", userphone+sn2order.OutTradeNo, "开始更新", sn2order.LogId)
				}
			}
		}
	}
	//结束更新现金账户与优惠券账户
	sn2order.OrderStatus = 7
	sn2order.ModifyTime = time.Now()
	err = l.svcCtx.UserOrder.Update(l.ctx, sn2order)
	sn, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	return &types.CancelOrderResp{Code: "10000", Msg: "发起退款成功", Data: &types.CancelOrderRp{OrderInfo: OrderDb2info(sn, info)}}, nil
}
