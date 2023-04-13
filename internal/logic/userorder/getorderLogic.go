package userorder

import (
	"context"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetorderLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	checklogic *CheckOrderLogic
	userphone  string
	useropenid string
}

func NewGetorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetorderLogic {
	return &GetorderLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		checklogic: NewCheckOrderLogic(ctx, svcCtx),
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
	}
}
func (l *GetorderLogic) Getorder(req *types.GetOrderRes) (resp *types.GetOrderResp, err error) {
	order, _ := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	transactioninfo, _ := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil || transactioninfo == nil {
		return &types.GetOrderResp{Code: "4004", Msg: "数据库失效"}, nil
	}
	if order.Phone != l.userphone {
		return &types.GetOrderResp{Code: "4004", Msg: "不要使用别人的token"}, nil
	}
	if order.OrderStatus == 0 {
		total, cash, weixin := IfFinished(transactioninfo)
		if total {
			return l.check01(order, transactioninfo)
		}
		if cash && !weixin && l.checklogic.CheckWeiXinPay(order.OutTradeNo) {
			l.svcCtx.TransactionInfo.UpdateWeixinPay(l.ctx, transactioninfo.OrderSn)
			return l.check01(order, transactioninfo)
		}
		return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(order, transactioninfo)}}, nil

	}
	if order.OrderStatus == 6 {
		total, cash, weixin := IfRejected(transactioninfo)
		if total {
			return l.check67(order, transactioninfo)
		}
		if cash && !weixin && l.checklogic.CheckWeiXinReject(order) {
			l.svcCtx.TransactionInfo.UpdateWeixinReject(l.ctx, transactioninfo.OrderSn)
			return l.check01(order, transactioninfo)
		}
	}
	return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(order, transactioninfo)}}, nil

}
func (l *GetorderLogic) check01(order *cachemodel.UserOrder, transactioninfo *cachemodel.TransactionInfo) (*types.GetOrderResp, error) {
	order.OrderStatus = 1
	order.WexinPayAmount = transactioninfo.WexinPayAmount
	order.CashAccountPayAmount = transactioninfo.CashAccountPayAmount
	order.FinishWeixinpay = transactioninfo.FinishWeixinpay
	order.PaymentTime = time.Now()
	order.ModifyTime = time.Now()
	l.svcCtx.UserOrder.Update(l.ctx, order)
	return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(order, nil)}}, nil
}
func (l *GetorderLogic) check67(order *cachemodel.UserOrder, transactioninfo *cachemodel.TransactionInfo) (*types.GetOrderResp, error) {
	order.OrderStatus = 7
	order.WexinPayAmount = transactioninfo.WexinPayAmount
	order.CashAccountPayAmount = transactioninfo.CashAccountPayAmount
	order.FinishWeixinpay = transactioninfo.FinishWeixinpay
	order.CloseTime = time.Now()
	order.ModifyTime = time.Now()
	l.svcCtx.UserOrder.Update(l.ctx, order)
	return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(order, nil)}}, nil
}
