package userorder

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
	wechatutil *WeChatUtilLogic
}

func NewRefundorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundorderLogic {
	return &RefundorderLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
		wechatutil: NewWeChatUtilLogic(ctx, svcCtx),
	}
}

func (l *RefundorderLogic) Refundorder(req *types.CancelOrderRes) (resp *types.CancelOrderResp, err error) {
	//必须注意，这个接口是发起退款接口，不参与判定是否退款成功
	lu := NewLogic(l.ctx, l.svcCtx)
	order, _ := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	transactioninfo, _ := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil || transactioninfo == nil {
		return &types.CancelOrderResp{Code: "4004", Msg: "未查询到订单"}, nil
	}
	if order.OrderStatus != 1 {
		return &types.CancelOrderResp{Code: "10000", Msg: "您好，您的订单非已付款未发货状态，暂不可取消订单，若商品有质量问题请点击投诉按钮投诉，我们会在24小时内为您处理！"}, nil
	}
	order.OrderStatus = 6 // 删除订单，设置为开始退款。
	order.WexinPayAmount = transactioninfo.WexinPayAmount
	order.CashAccountPayAmount = transactioninfo.CashAccountPayAmount
	order.ModifyTime = time.Now()
	l.svcCtx.UserOrder.Update(l.ctx, order)

	// 首先开始退微信支付的钱
	if order.WexinPayAmount > 0 {
		l.wechatutil.CancelOrder(order)
	}
	// 然后开始退现金账户的钱和优惠券
	if order.CashAccountPayAmount > 0 || order.UsedCouponinfo != "" {
		lockmsglist := make([]*types.LockMsg, 0)
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.userphone, Field: "user_coupon"})
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.userphone, Field: "cash_account"})
		if lu.Getlocktry(lockmsglist) {
			if order.CashAccountPayAmount > 0 {
				ok, _ := lu.Updatecashaccount(order, false)
				if !ok {
					lu.Oplog("更新现金账户失败", order.OrderSn, "开始更新", order.LogId)
				}
			}
			if order.UsedCouponinfo != "" {
				ok, _ := lu.UpdateCoupon(order, false)
				if !ok {
					lu.Oplog("更新优惠券失败", order.OrderSn, "开始更新", order.LogId)
				}
			}
		}

		lu.Closelock(lockmsglist)
	}
	//结束更新现金账户与优惠券账户
	// 开始退积分账户
	if order.PointAmount > 0 {
		userpoints, _ := l.svcCtx.UserPoints.FindOneByPhone(l.ctx, order.Phone)
		if userpoints != nil {
			userpoints.AvailablePoints = userpoints.AvailablePoints + order.PointAmount
			l.svcCtx.UserPoints.Update(l.ctx, userpoints)
		}
	}
	return &types.CancelOrderResp{Code: "10000", Msg: "yes", Data: &types.CancelOrderRp{OrderInfo: OrderDb2info(order, transactioninfo)}}, nil
}
