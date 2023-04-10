package userorder

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ContinuepayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewContinuepayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ContinuepayLogic {
	return &ContinuepayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ContinuepayLogic) Continuepay(req *types.FinishOrderRes) (resp *types.NewOrderResp, err error) {
	UseAccount := false
	order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil {
		return &types.NewOrderResp{Code: "4004", Msg: "未查询到订单,请重建订单"}, nil
	}
	if order.OrderStatus == 1 {
		return &types.NewOrderResp{Code: "4004", Msg: "此订单已支付"}, nil
	}
	payl := NewPayLogic(l.ctx, l.svcCtx)
	payorder, success := payl.Payorder(&types.TransactionInit{TransactionType: "普通商品", OrderSn: order.OrderSn, NeedCashAccount: true, Ammount: order.ActualAmount, Phone: order.Phone})
	if !success {
		return &types.NewOrderResp{Code: "4004", Msg: "fatal error"}, nil
	}
	if order.PointAmount > 0 || order.UsedCouponinfo != "" || payorder.NeedCashAccountPay {
		UseAccount = true
	}
	order.WexinPayAmount = payorder.WeiXinPayAmmount
	order.CashAccountPayAmount = payorder.CashPayAmmount
	l.svcCtx.UserOrder.Update(l.ctx, order)
	neworderrp := types.NewOrderRp{OrderInfo: OrderDb2info(order, nil), UseAccount: UseAccount, UseWechatPay: payorder.NeedWeiXinPay, WeiXinPayMsg: payorder.WeiXinPayMsg}
	return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &neworderrp}, nil
}
