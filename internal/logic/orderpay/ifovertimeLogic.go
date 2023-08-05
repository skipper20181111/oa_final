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
		Msg:  "success",
		Data: &types.IfOvertimeRp{
			OverTimeMilliSecondsList: make([]int64, 0),
		},
	}
	for _, OrderSn := range req.OrderSnList {
		overTime, ok := l.GetOverTime(OrderSn)
		if !ok {
			resp.Msg = "Not All Overed,or Db Error"
		}
		resp.Data.OverTimeMilliSecondsList = append(resp.Data.OverTimeMilliSecondsList, overTime)
	}
	return resp, nil
}
func (l IfovertimeLogic) GetOverTime(OrderSn string) (int64, bool) {
	OverTime := int64(1000)
	order, _ := l.svcCtx.Order.FindOneByOrderSn(l.ctx, OrderSn)
	if order == nil {
		return OverTime, false
	} else {
		OverTime = order.CreateOrderTime.Add(time.Minute*15).UnixMilli() - time.Now().UnixMilli()
	}
	payinfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)
	if payinfo == nil {
		return OverTime, false
	}
	if OrderCanBeOvertime(order, payinfo) {
		l.svcCtx.Order.UpdateStatusByOrderSn(l.ctx, 8, order.OrderSn)
		return OverTime, true
	}
	return OverTime, false
}
