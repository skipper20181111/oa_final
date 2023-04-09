package userorder

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"log"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetorderLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
}

func NewGetorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetorderLogic {
	return &GetorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (l *GetorderLogic) Getorder(req *types.GetOrderRes) (resp *types.GetOrderResp, err error) {
	l.userphone = l.ctx.Value("phone").(string)
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	sn, err := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if sn2order == nil || sn == nil {
		return &types.GetOrderResp{Code: "4004", Msg: err.Error()}, nil
	}
	if sn2order.Phone != l.userphone {
		return &types.GetOrderResp{Code: "4004", Msg: "不要使用别人的token"}, nil
	}
	if IfFinished(sn) {
		if sn2order.OrderStatus == 0 {
			sn2order.OrderStatus = 1
			sn2order.WexinPayAmount = sn.WexinPayAmount
			sn2order.CashAccountPayAmount = sn.CashAccountPayAmount
			l.svcCtx.UserOrder.Update(l.ctx, sn2order)
		}
		return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(sn2order, nil)}}, nil
	} else {
		jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
		no2payment, result, err := jssvc.QueryOrderByOutTradeNo(l.ctx, jsapi.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(sn2order.OutTradeNo),
			Mchid:      core.String(l.svcCtx.Config.WxConf.MchID)})
		defer result.Response.Body.Close()
		if err != nil {
			return &types.GetOrderResp{Code: "4004", Msg: err.Error()}, nil
		}
		if *no2payment.TradeState != "SUCCESS" {
			l.svcCtx.TransactionInfo.UpdateWeixinPay(l.ctx, sn.Phone)
			sn.FinishWeixinpay = 1
			if IfFinished(sn) {
				sn2order.OrderStatus = 1
				sn2order.WexinPayAmount = sn.WexinPayAmount
				sn2order.CashAccountPayAmount = sn.CashAccountPayAmount
				l.svcCtx.UserOrder.Update(l.ctx, sn2order)
			}
			return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(sn2order, nil)}}, nil
		}
	}
	if sn2order.OrderStatus == 6 { // 说明已经发起了退款，具体有没有成功呢？
		service := refunddomestic.RefundsApiService{Client: l.svcCtx.Client}
		no, result, err := service.QueryByOutRefundNo(l.ctx,
			refunddomestic.QueryByOutRefundNoRequest{
				OutRefundNo: core.String(sn2order.OutTradeNo),
			})
		defer result.Response.Body.Close()
		if err != nil {
			log.Printf("call QueryByOutRefundNo err:%s", err)
			return &types.GetOrderResp{Code: "4004", Msg: err.Error()}, nil

		} else {
			log.Printf("status=%d resp=%s", result.Response.StatusCode, resp)
		}
		if *no.Status == "SUCCESS" {
			sn2order.OrderStatus = 7
			sn2order.ModifyTime = time.Now()
			err := l.svcCtx.UserOrder.Update(l.ctx, sn2order)
			if err != nil {
				fmt.Println(err.Error())
			}
			sn2order, err = l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, sn2order.OrderSn)
			if err != nil {
				fmt.Println(err.Error())
			}
			return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(sn2order, nil)}}, nil

		}
	}
	return &types.GetOrderResp{Code: "10000", Msg: "查询成功", Data: &types.GetOrderRp{OrderInfo: OrderDb2info(sn2order, nil)}}, nil

}
