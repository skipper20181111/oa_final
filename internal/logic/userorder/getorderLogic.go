package userorder

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetorderLogic {
	return &GetorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetorderLogic) Getorder(req *types.GetOrderRes) (resp *types.GetOrderResp, err error) {
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	if err != nil {
		return &types.GetOrderResp{Code: "4004", Msg: err.Error()}, nil
	}
	if sn2order.OrderStatus == 0 { // 说明还没有付款，去查一查究竟有没有付款
		jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
		no2payment, _, err := jssvc.QueryOrderByOutTradeNo(l.ctx, jsapi.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String("PIS5TwTkh3Z0vv5G3BzT2NLorgVQi52P"),
			Mchid:      core.String(l.svcCtx.Config.WxConf.MchID)})
		if err != nil {
			return &types.GetOrderResp{Code: "4004", Msg: err.Error()}, nil
		}
		if *no2payment.TradeState != "SUCCESS" {
			return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: db2orderinfo(sn2order)}}, nil
		} else {
			sn2order.OrderStatus = 1
			sn2order.ModifyTime = time.Now()
			sn2order.PaymentTime = sn2order.ModifyTime
			err := l.svcCtx.UserOrder.Update(l.ctx, sn2order)
			if err != nil {
				fmt.Println(err.Error())
			}
			sn, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, sn2order.OrderSn)
			if err != nil {
				fmt.Println(err.Error())
			}
			return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: db2orderinfo(sn)}}, nil
		}
	}

	return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: db2orderinfo(sn2order)}}, nil

}
