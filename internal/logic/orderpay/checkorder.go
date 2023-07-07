package orderpay

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"time"
)

type CheckOrderLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	userphone  string
	useropenid string
	wul        *WeChatUtilLogic
}

func NewCheckOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckOrderLogic {
	return &CheckOrderLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
		wul:        NewWeChatUtilLogic(ctx, svcCtx),
	}
}
func (l *CheckOrderLogic) MonitorOrderStatus(OrderSn string) (*types.GetOrderResp, error) {
	order, _ := l.svcCtx.Order.FindOneByOrderSn(l.ctx, OrderSn)
	if order == nil {
		return &types.GetOrderResp{Code: "4004", Msg: "数据库失效，请重新下单"}, nil
	}
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)
	if PayInfo == nil {
		return &types.GetOrderResp{Code: "4004", Msg: "数据库失效，请重新下单"}, nil
	}
	if order.Phone != l.userphone {
		return &types.GetOrderResp{Code: "4004", Msg: "不要使用别人的token"}, nil
	}
	order = l.checkall(order, PayInfo)
	return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(order)}}, nil
}

func (l *CheckOrderLogic) checkall(order *cachemodel.Order, PayInfo *cachemodel.PayInfo) *cachemodel.Order {
	if PayInfo == nil {
		return order
	}
	if order.OrderStatus == 0 {
		total, cash, weixin := IfFinished(PayInfo)
		if total {
			return l.check01(order, PayInfo)
		}
		finishwechatpay := !weixin && l.wul.CheckWeiXinPay(order.OutTradeNo)
		if cash && finishwechatpay {
			return l.check01(order, PayInfo)
		}
		return order
	}
	if order.OrderStatus == 6 {
		total, cash, weixin := IfRejected(order)
		if total {
			return l.check67(order, PayInfo)
		}
		if cash && !weixin && l.wul.IfCancelOrderSuccess(order) {
			return l.check67(order, PayInfo)
		}
		return order
	}
	return order
}
func IfFinished(PayInfo *cachemodel.PayInfo) (total bool, cash bool, weixin bool) {
	total = false
	cash = false
	weixin = false
	if PayInfo.CashAccountPayAmount > 0 && PayInfo.FinishAccountpay == 1 {
		cash = true
	}
	if PayInfo.CashAccountPayAmount <= 0 {
		cash = true
	}
	if PayInfo.WexinPayAmount > 0 && PayInfo.FinishWeixinpay == 1 {
		weixin = true
	}
	if PayInfo.WexinPayAmount <= 0 {
		weixin = true
	}
	return weixin && cash, cash, weixin
}
func (l *CheckOrderLogic) check01(order *cachemodel.Order, PayInfo *cachemodel.PayInfo) *cachemodel.Order {
	order.OrderStatus = 1
	order.FinishWeixinpay = PayInfo.FinishWeixinpay
	order.PaymentTime = time.Now()
	order.ModifyTime = time.Now()
	l.svcCtx.PayInfo.UpdateAllPay(l.ctx, order.OrderSn)
	l.svcCtx.Order.Update(l.ctx, order)
	return order
}
func IfRejected(order *cachemodel.Order) (total bool, cash bool, weixin bool) {
	total = false
	cash = false
	weixin = false
	if order.CashAccountPayAmount > 0 && order.FinishAccountpay == -1 {
		cash = true
	}
	if order.CashAccountPayAmount <= 0 {
		cash = true
	}
	if order.WexinPayAmount > 0 && order.FinishWeixinpay == -1 {
		weixin = true
	}
	if order.WexinPayAmount <= 0 {
		weixin = true
	}
	return weixin && cash, cash, weixin
}
func (l *CheckOrderLogic) check67(order *cachemodel.Order, PayInfo *cachemodel.PayInfo) *cachemodel.Order {
	order.OrderStatus = 7
	order.CloseTime = time.Now()
	order.ModifyTime = time.Now()
	l.svcCtx.Order.Update(l.ctx, order)
	return order
}
