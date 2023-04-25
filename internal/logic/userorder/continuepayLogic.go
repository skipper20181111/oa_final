package userorder

import (
	"context"
	"oa_final/cachemodel"

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
	PMcache, ok := l.svcCtx.LocalCache.Get(svc.ProductsMap)
	if !ok {
		return &types.NewOrderResp{Code: "4004", Msg: "服务器查找商品列表失败"}, nil
	}
	productsMap := PMcache.(map[int64]*cachemodel.Product)
	lu := NewLogic(l.ctx, l.svcCtx)
	neworder := lu.Order2db(order2req(order), productsMap, UseCache(false))
	neworder.Id = order.Id
	neworder.Address = order.Address
	neworder.OrderSn = order.OrderSn

	neworder.LogId = order.LogId
	payl := NewPayLogic(l.ctx, l.svcCtx)
	l.svcCtx.UserOrder.Update(l.ctx, neworder)
	payorder, success := payl.Payorder(&types.TransactionInit{TransactionType: "普通商品", OrderSn: order.OrderSn, OutTradeSn: neworder.OutTradeNo, NeedCashAccount: true, Ammount: order.ActualAmount, Phone: order.Phone})
	if !success {
		return &types.NewOrderResp{Code: "4004", Msg: "fatal error"}, nil
	}
	if neworder.PointAmount > 0 || neworder.UsedCouponinfo != "" || payorder.NeedCashAccountPay {
		UseAccount = true
	}
	neworder.WexinPayAmount = payorder.WeiXinPayAmmount
	neworder.CashAccountPayAmount = payorder.CashPayAmmount
	l.svcCtx.UserOrder.Update(l.ctx, neworder)
	neworderrp := types.NewOrderRp{OrderInfo: OrderDb2info(neworder, nil), UseAccount: UseAccount, UseWechatPay: payorder.NeedWeiXinPay, WeiXinPayMsg: payorder.WeiXinPayMsg}
	return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &neworderrp}, nil
}
