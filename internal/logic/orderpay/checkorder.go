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
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
	wul       *WeChatUtilLogic
	u         *UtilLogic
}

func NewCheckOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckOrderLogic {
	return &CheckOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		wul:    NewWeChatUtilLogic(ctx, svcCtx),
		u:      NewUtilLogic(ctx, svcCtx),
	}
}
func (l *CheckOrderLogic) MonitorOrderStatus(OutTradeNo string) (*types.GetAllOrderResp, error) {
	l.userphone = l.ctx.Value("phone").(string)
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, OutTradeNo)
	if PayInfo == nil {
		return &types.GetAllOrderResp{Code: "4004", Msg: "数据库失效，请重新下单"}, nil
	}
	if PayInfo.Phone != l.userphone {
		return &types.GetAllOrderResp{Code: "4004", Msg: "不要使用别人的token"}, nil
	}
	OrderInfos := make([]*types.OrderInfo, 0)
	Orders, _ := l.svcCtx.Order.FindAllByOutTradeNo(l.ctx, OutTradeNo)
	for _, order := range Orders {
		OrderInfos = append(OrderInfos, l.u.OrderDb2info(l.Checkall(order, PayInfo)))
	}
	return &types.GetAllOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetAllOrderRp{OrderInfos: OrderInfos}}, nil
}

func (l *CheckOrderLogic) Checkall(order *cachemodel.Order, PayInfo *cachemodel.PayInfo) *cachemodel.Order {
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
	if PayInfo.Status == 1 {
		return true, true, true
	}
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
	l.svcCtx.PayInfo.UpdateAllPay(l.ctx, order.OutTradeNo)
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
