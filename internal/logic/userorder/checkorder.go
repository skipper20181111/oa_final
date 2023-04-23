package userorder

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"time"
)

type CheckOrderLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	userphone       string
	useropenid      string
	WeChatUtilLogic *WeChatUtilLogic
}

func NewCheckOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckOrderLogic {
	return &CheckOrderLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		userphone:       ctx.Value("phone").(string),
		useropenid:      ctx.Value("openid").(string),
		WeChatUtilLogic: NewWeChatUtilLogic(ctx, svcCtx),
	}
}

func (l *CheckOrderLogic) Payorder(req *types.TransactionInit) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
}
func (l *CheckOrderLogic) CheckWeiXinPay(OutTradeNo string) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	no2payment, result, _ := jssvc.QueryOrderByOutTradeNo(l.ctx, jsapi.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(OutTradeNo),
		Mchid:      core.String(l.svcCtx.Config.WxConf.MchID)})
	defer result.Response.Body.Close()
	if *no2payment.TradeState != "SUCCESS" {
		return true
	}
	return false
}
func (l *CheckOrderLogic) CheckWeiXinReject(Order *cachemodel.UserOrder) bool {
	return l.WeChatUtilLogic.IfCancelOrderSuccess(Order)
}
func OrderCanBeDeleted(order *cachemodel.UserOrder, transactioninfo *cachemodel.TransactionInfo) bool {
	if order.OrderStatus == 7 || order.OrderStatus == 3 {
		return true
	}
	if order.OrderStatus == 0 {
		if !PartPay(order, transactioninfo) {
			return true
		} else {
			return false
		}
	}
	return false
}
func PartPay(order *cachemodel.UserOrder, info *cachemodel.TransactionInfo) bool {
	cash := false
	weixin := false
	if info.CashAccountPayAmount > 0 && info.FinishAccountpay == 1 {
		cash = true
	}
	if info.WexinPayAmount > 0 && info.FinishWeixinpay == 1 {
		weixin = true
	}
	if cash == false && weixin == false {
		return false
	} else {
		return true
	}
}
func OrderNeedChange(order *cachemodel.UserOrder) bool {
	if order.OrderStatus == 0 || order.OrderStatus == 6 {
		return true
	} else {
		return false
	}
}
func (l *CheckOrderLogic) checkall(order *cachemodel.UserOrder, transactioninfo *cachemodel.TransactionInfo) *cachemodel.UserOrder {
	if transactioninfo == nil {
		return order
	}
	if order.OrderStatus == 0 {
		total, cash, weixin := IfFinished(transactioninfo)
		if total {
			return l.check01(order, transactioninfo)
		}
		if cash && !weixin && l.CheckWeiXinPay(order.OutTradeNo) {
			l.svcCtx.TransactionInfo.UpdateWeixinPay(l.ctx, transactioninfo.OrderSn)
			order = l.check01(order, transactioninfo)
			return l.check01(order, transactioninfo)
		}
		return order
	}
	if order.OrderStatus == 6 {
		total, cash, weixin := IfRejected(transactioninfo)
		if total {
			return l.check67(order, transactioninfo)

		}
		if cash && !weixin && l.CheckWeiXinReject(order) {
			l.svcCtx.TransactionInfo.UpdateWeixinReject(l.ctx, transactioninfo.OrderSn)
			order = l.check67(order, transactioninfo)
			return l.check67(order, transactioninfo)
		}
		return order
	}
	return order
}
func (l *CheckOrderLogic) MonitorOrderStatus(OrderSn string) (*types.GetOrderResp, error) {
	order, _ := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, OrderSn)
	transactioninfo, _ := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, OrderSn)
	if order == nil || transactioninfo == nil {
		return &types.GetOrderResp{Code: "10000", Msg: "此订单不存在"}, nil
	}
	if order.Phone != l.userphone {
		return &types.GetOrderResp{Code: "4004", Msg: "不要使用别人的token"}, nil
	}
	order = l.checkall(order, transactioninfo)
	return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(order, transactioninfo)}}, nil
}
func (l *CheckOrderLogic) check01(order *cachemodel.UserOrder, transactioninfo *cachemodel.TransactionInfo) *cachemodel.UserOrder {
	order.OrderStatus = 1
	order.WexinPayAmount = transactioninfo.WexinPayAmount
	order.CashAccountPayAmount = transactioninfo.CashAccountPayAmount
	order.FinishWeixinpay = transactioninfo.FinishWeixinpay
	order.PaymentTime = time.Now()
	order.ModifyTime = time.Now()
	l.svcCtx.UserOrder.Update(l.ctx, order)
	return order
}
func (l *CheckOrderLogic) check67(order *cachemodel.UserOrder, transactioninfo *cachemodel.TransactionInfo) *cachemodel.UserOrder {
	order.OrderStatus = 7
	order.WexinPayAmount = transactioninfo.WexinPayAmount
	order.CashAccountPayAmount = transactioninfo.CashAccountPayAmount
	order.FinishWeixinpay = transactioninfo.FinishWeixinpay
	order.CloseTime = time.Now()
	order.ModifyTime = time.Now()
	l.svcCtx.UserOrder.Update(l.ctx, order)
	return order
}
func IfFinished(info *cachemodel.TransactionInfo) (total bool, cash bool, weixin bool) {
	total = false
	cash = false
	weixin = false
	if info.NeedCashAccount == 1 && info.FinishAccountpay == 1 {
		cash = true
	}
	if info.NeedCashAccount == 0 {
		cash = true
	}
	if info.WexinPayAmount > 0 && info.FinishWeixinpay == 1 {
		weixin = true
	}
	if info.WexinPayAmount <= 0 {
		weixin = true
	}
	return weixin && cash, cash, weixin
}
func IfRejected(info *cachemodel.TransactionInfo) (total bool, cash bool, weixin bool) {
	total = false
	cash = false
	weixin = false
	if info.NeedCashAccount == 1 && info.FinishAccountpay == -1 {
		cash = true
	}
	if info.NeedCashAccount == 0 {
		cash = true
	}
	if info.WexinPayAmount > 0 && info.FinishWeixinpay == -1 {
		weixin = true
	}
	if info.WexinPayAmount <= 0 {
		weixin = true
	}
	return weixin && cash, cash, weixin
}
