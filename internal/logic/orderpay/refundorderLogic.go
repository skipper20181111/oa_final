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
	//必须注意，这个接口是发起退款接口，不参与判定是否退款成功
	order, _ := l.svcCtx.Order.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil {
		return &types.CancelOrderResp{Code: "4004", Msg: "数据库失效，请重新下单"}, nil
	}
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)
	if PayInfo == nil {
		return &types.CancelOrderResp{Code: "4004", Msg: "数据库失效，请重新下单"}, nil
	}
	if order.OrderStatus != 1 || PayInfo.Status != 1 {
		return &types.CancelOrderResp{Code: "10000", Msg: "您好，您的订单非已付款未发货状态，暂不可取消订单，若商品有质量问题请点击投诉按钮投诉，我们会在24小时内为您处理！"}, nil
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
	return &types.CancelOrderResp{Code: "10000", Msg: "yes", Data: &types.CancelOrderRp{OrderInfo: OrderDb2info(order)}}, nil

}
