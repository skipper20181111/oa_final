package orderpay

import (
	"context"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type IfovertimeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIfovertimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IfovertimeLogic {
	return &IfovertimeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IfovertimeLogic) Ifovertime(req *types.IfOvertimeRes) (resp *types.IfOvertimeResp, err error) {
	resp = &types.IfOvertimeResp{
		Code: "10000",
		Msg:  "NoGood",
		Data: &types.IfOvertimeRp{},
	}
	order, _ := l.svcCtx.Order.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil {
		return resp, nil
	} else {
		resp.Data.OverTimeMilliSeconds = order.CreateOrderTime.Add(time.Minute*15).UnixMilli() - time.Now().UnixMilli()
	}
	payinfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)
	if payinfo == nil {
		return resp, nil
	}
	if OrderCanBeOvertime(order, payinfo) {
		l.svcCtx.Order.UpdateStatusByOrderSn(l.ctx, 8, order.OrderSn)
		resp.Msg = "success"
		return resp, nil
	}
	return resp, nil
}
