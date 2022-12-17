package userorder

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"log"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelorderLogic {
	return &CancelorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelorderLogic) Cancelorder(req *types.CancelOrderRes) (resp *types.CancelOrderResp, err error) {
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	if err != nil {
		return &types.CancelOrderResp{Code: "4004", Msg: err.Error()}, nil
	}
	if sn2order.OrderStatus != 1 {
		return &types.CancelOrderResp{Code: "4004", Msg: "无法退款"}, nil
	}
	service := refunddomestic.RefundsApiService{Client: l.svcCtx.Client}
	create, result, err := service.Create(l.ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(sn2order.OutTradeNo),
		OutRefundNo: core.String(sn2order.OutTradeNo),
		Amount: &refunddomestic.AmountReq{Currency: core.String("CNY"),
			Refund: core.Int64(1),
			Total:  core.Int64(1)},
	})
	defer result.Response.Body.Close()
	if err != nil {
		log.Printf("call Create err:%s", err)
		return &types.CancelOrderResp{Code: "4004", Msg: err.Error()}, nil
	} else {
		log.Printf("status=%d resp=%s", result.Response.StatusCode, resp, create.String())
	}
	sn2order.OrderStatus = 6
	sn2order.ModifyTime = time.Now()
	err = l.svcCtx.UserOrder.Update(l.ctx, sn2order)
	sn, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	return &types.CancelOrderResp{Code: "10000", Msg: "发起退款成功", Data: &types.CancelOrderRp{OrderInfo: db2orderinfo(sn)}}, nil
}
