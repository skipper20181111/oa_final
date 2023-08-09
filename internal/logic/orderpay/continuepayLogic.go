package orderpay

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
	pu     *PayUtilLogic
	ou     *OrderUtilLogic
}

func NewContinuepayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ContinuepayLogic {
	return &ContinuepayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		pu:     NewPayUtilLogic(ctx, svcCtx),
		ou:     NewOrderUtilLogic(ctx, svcCtx),
	}
}

func (l *ContinuepayLogic) Continuepay(req *types.ContinuePayRes) (resp *types.NewOrderResp, err error) {
	PayInfo, err := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, req.OutTradeNo)
	if PayInfo == nil {
		return &types.NewOrderResp{Code: "4004", Msg: "未查询到订单,请重建订单"}, nil
	}
	if PayInfo.Status != 0 {
		return &types.NewOrderResp{Code: "4004", Msg: "此订单不可支付"}, nil
	}
	UseAccount := false
	UseCoupon := false
	OrderInfos := make([]*types.OrderInfo, 0)
	NewOrderRes := PayInfo2req(PayInfo, req)
	OrderList, payInit, ok := l.ou.req2op(NewOrderRes)
	if !ok {
		return &types.NewOrderResp{Code: "10000", Msg: "error", Data: &types.NewOrderRp{}}, nil
	}
	payInit.TransactionType = "普通商品"
	payMsg, orders, payinfo, success := l.pu.Payorder(payInit, OrderList)
	if !success {
		return &types.NewOrderResp{Code: "4004", Msg: "fatal error"}, nil
	}
	l.svcCtx.Order.DeleteByOutTradeSn(l.ctx, payInit.OutTradeSn)
	l.svcCtx.PayInfo.Delete(l.ctx, PayInfo.Id)
	for _, order := range orders {
		order.OutTradeNo = payInit.OutTradeSn
		l.svcCtx.Order.Insert(l.ctx, order)
		OrderInfos = append(OrderInfos, OrderDb2info(order))
		if order.UsedCouponinfo != "" {
			UseCoupon = true
		}
	}
	if payMsg.CashPayAmmount > 0 || UseCoupon {
		UseAccount = true
	}
	neworderrp := types.NewOrderRp{PayInfo: payinfo, OrderInfos: OrderInfos, UseAccount: UseAccount, UseWechatPay: payMsg.NeedWeChatPay, WeiXinPayMsg: payMsg.WeChatPayMsg}
	return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &neworderrp}, nil
}
