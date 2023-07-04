package orderpay

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteorderLogic {
	return &DeleteorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteorderLogic) Deleteorder(req *types.DeletOrderRes) (resp *types.DeletOrderResp, err error) {
	order, _ := l.svcCtx.Order.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil {
		return &types.DeletOrderResp{Code: "10000", Msg: "订单不存在"}, nil
	}
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)

	if PayInfo == nil {
		l.svcCtx.Order.Delete(l.ctx, order.Id)
		return &types.DeletOrderResp{Code: "10000", Msg: "yes"}, nil
	}
	if OrderCanBeDeleted(order, PayInfo) {
		l.svcCtx.Order.UpdateStatusByOrderSn(l.ctx, 9, order.OrderSn)
		return &types.DeletOrderResp{Code: "10000", Msg: "yes"}, nil
	}
	return &types.DeletOrderResp{Code: "10000", Msg: "订单状态不可删除"}, nil
}
