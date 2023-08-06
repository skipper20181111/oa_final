package orderpay

import (
	"context"
	"oa_final/cachemodel"
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
			OverTimeMilliSecondsMap: make(map[string]int64, 0),
		},
	}
	payinfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, req.OutTradeSn)
	if payinfo == nil {
		resp.Code = "4004"
		return resp, nil
	}
	orders, _ := l.svcCtx.Order.FindAllByOutTradeNo(l.ctx, req.OutTradeSn)
	for _, order := range orders {
		overTime, ok := l.GetOverTime(order, payinfo)
		if !ok {
			resp.Msg = "Not All Overed,or Db Error"
		}
		resp.Data.OverTimeMilliSecondsMap[order.OrderSn] = overTime
	}

	return resp, nil
}
func (l IfovertimeLogic) GetOverTime(Order *cachemodel.Order, payinfo *cachemodel.PayInfo) (int64, bool) {
	//OverTime := Order.CreateOrderTime.Add(time.Minute*15).UnixMilli() - time.Now().UnixMilli()

	OverTime := Order.CreateOrderTime.Add(time.Second*30).UnixMilli() - time.Now().UnixMilli()
	if OrderCanBeOvertime(Order, payinfo) {
		l.svcCtx.Order.UpdateStatusByOrderSn(l.ctx, 8, Order.OrderSn)
		return OverTime, true
	}
	return OverTime, false
}
