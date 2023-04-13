package userorder

import (
	"context"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelorderLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	userphone  string
	useropenid string
	wechatutil *WeChatUtilLogic
}

func NewCancelorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelorderLogic {
	return &CancelorderLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
		wechatutil: NewWeChatUtilLogic(ctx, svcCtx),
	}
}

func (l *CancelorderLogic) Cancelorder(req *types.CancelOrderRes) (resp *types.CancelOrderResp, err error) {
	//必须注意，这个接口是发起退款接口，不参与判定是否退款成功
	lu := NewLogic(l.ctx, l.svcCtx)
	order, _ := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	transactioninfo, _ := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil || transactioninfo == nil {
		return &types.CancelOrderResp{Code: "4004", Msg: "未查询到订单"}, nil
	}
	if order.OrderStatus != 1 {
		return &types.CancelOrderResp{Code: "10000", Msg: "您好，您订购的商品已进入运输环节，暂不可取消订单，若商品有质量问题请点击投诉按钮投诉，我们会在24小时内为您处理！"}, nil
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
		if lu.getlocktry(lockmsglist) {
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
		lu.closelock(lockmsglist)
	}
	//结束更新现金账户与优惠券账户
	return &types.CancelOrderResp{Code: "10000", Msg: "发起退款成功", Data: &types.CancelOrderRp{OrderInfo: OrderDb2info(order, transactioninfo)}}, nil
}
