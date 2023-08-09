package deliver

import (
	"context"
	"oa_final/internal/logic/orderpay"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreparegoodsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	sf     *orderpay.SfUtilLogic
}

func NewPreparegoodsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreparegoodsLogic {
	return &PreparegoodsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    context.Background(),
		svcCtx: svcCtx,
		sf:     orderpay.NewSfUtilLogic(context.Background(), svcCtx),
	}
}

func (l *PreparegoodsLogic) Preparegoods(req *types.PrepareGoodsRes) (resp *types.PrepareGoodsResp, err error) {
	resp = &types.PrepareGoodsResp{
		Code: "10000",
		Msg:  "success",
		Data: &types.PrepareGoodsRp{
			SuccessOrderSn: make([]string, 0),
			FailedOrderSn:  make([]string, 0),
		},
	}
	for _, orderSn := range req.OrderSns {
		if l.ChangeOrderStatus(orderSn) {
			resp.Data.SuccessOrderSn = append(resp.Data.SuccessOrderSn, orderSn)
		} else {
			resp.Data.FailedOrderSn = append(resp.Data.FailedOrderSn, orderSn)
		}
	}
	return resp, nil
}
func (l PreparegoodsLogic) ChangeOrderStatus(OrderSn string) bool {
	order, _ := l.svcCtx.Order.FindOneByOrderSn(l.ctx, OrderSn)
	if order.OrderStatus == 1 {
		l.svcCtx.Order.UpdateStatusByOrderSn(l.ctx, 1001, OrderSn)
		order, _ = l.svcCtx.Order.FindOneByOrderSn(l.ctx, OrderSn)
		if order.OrderStatus == 1001 {
			l.sf.GetSfSn(order)
			return true
		}
	}
	return false
}
