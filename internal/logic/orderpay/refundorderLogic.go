package orderpay

import (
	"context"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefundorderLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	userphone  string
	useropenid string
	wcu        *WeChatUtilLogic
	ul         *UtilLogic
	oul        *OrderUtilLogic
}

func NewRefundorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundorderLogic {
	return &RefundorderLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
		wcu:        NewWeChatUtilLogic(ctx, svcCtx),
		ul:         NewUtilLogic(ctx, svcCtx),
		oul:        NewOrderUtilLogic(ctx, svcCtx),
	}
}

func (l *RefundorderLogic) Refundorder(req *types.CancelOrderRes) (resp *types.CancelOrderResp, err error) {
	resp = &types.CancelOrderResp{Code: "10000",
		Msg: "Success",
		Data: &types.CancelOrderRp{
			SuccessOrderInfos: make([]*types.OrderInfo, 0),
			FailedOrderInfos:  make([]string, 0),
		},
	}

	for _, OrderSn := range req.OrderSn {
		orderInfo, ok := l.RefundOneOrder(OrderSn)
		if !ok {
			resp.Data.FailedOrderInfos = append(resp.Data.FailedOrderInfos, OrderSn)
		} else {
			resp.Data.SuccessOrderInfos = append(resp.Data.SuccessOrderInfos, orderInfo)
		}
	}
	return resp, nil
}
func (l RefundorderLogic) RefundOneOrder(OrderSn string) (*types.OrderInfo, bool) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	//必须注意，这个接口是发起退款接口，不参与判定是否退款成功
	order, _ := l.svcCtx.Order.FindOneByOrderSn(l.ctx, OrderSn)
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)
	if order.OrderStatus != 1 || PayInfo.Status != 1 {
		return nil, false
	}
	order.OrderStatus = 6 // 删除订单，设置为开始退款。
	order.ModifyTime = time.Now()
	l.svcCtx.Order.Update(l.ctx, order)

	// 首先开始退微信支付的钱
	if order.WexinPayAmount > 0 {
		l.wcu.CancelOrder(order)
	}
	if (order.CashAccountPayAmount+PayInfo.CashAccountRefundAmount != PayInfo.CashAccountPayAmount) || (order.WexinPayAmount+PayInfo.WexinRefundAmount != PayInfo.WexinPayAmount) {
		order.UsedCouponinfo = ""
	}
	// 然后开始退现金账户的钱和优惠券
	if order.CashAccountPayAmount > 0 || len(order.UsedCouponinfo) > 6 {
		lockmsglist := make([]*types.LockMsg, 0)
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.userphone, Field: "user_coupon"})
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.userphone, Field: "cash_account"})
		if l.ul.Getlocktry(lockmsglist) {
			if order.CashAccountPayAmount > 0 {
				ok, _ := l.oul.Updatecashaccount(order.Phone, order.OrderSn, order.OutTradeNo, order.CashAccountPayAmount, order.LogId, false)
				if !ok {
					l.ul.Oplog("更新现金账户失败", order.OrderSn, "开始更新", order.LogId)
				}
			}
			if len(order.UsedCouponinfo) > 6 {
				ok, _ := l.oul.UpdateCoupon(order, false)
				if !ok {
					l.ul.Oplog("更新优惠券失败", order.OrderSn, "开始更新", order.LogId)
				}
			}
			l.ul.Closelock(lockmsglist)
		}

	}
	//结束更新现金账户与优惠券账户
	return OrderDb2info(order), true
}
