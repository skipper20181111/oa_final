package orderpay

import (
	"context"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type NeworderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	pu     *PayUtilLogic
	ou     *OrderUtilLogic
}

func NewNeworderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NeworderLogic {
	return &NeworderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		pu:     NewPayUtilLogic(ctx, svcCtx),
		ou:     NewOrderUtilLogic(ctx, svcCtx),
	}
}

func (l *NeworderLogic) Neworder(req *types.NewOrderRes) (resp *types.NewOrderResp, err error) {
	UseAccount := false
	UseCoupon := false
	OrderInfos := make([]*types.OrderInfo, 0)
	if len(req.ProductTinyList) == 0 {
		return &types.NewOrderResp{Code: "4004", Msg: "无商品，订单金额为0", Data: &types.NewOrderRp{}}, nil
	}
	OrderList, payInit, ok := l.ou.req2op(req)
	if !ok {
		return &types.NewOrderResp{Code: "10000", Msg: "error", Data: &types.NewOrderRp{}}, nil
	}
	payInit.TransactionType = "普通商品"
	payMsg, orders, payinfo, success := l.pu.Payorder(payInit, OrderList)
	if !success {
		return &types.NewOrderResp{Code: "4004", Msg: "fatal error"}, nil
	}
	for _, order := range orders {
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
