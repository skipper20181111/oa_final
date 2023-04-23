package userorder

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
	order, _ := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	transactionInfo, _ := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil {
		return &types.DeletOrderResp{Code: "10000", Msg: "订单不存在"}, nil
	}
	if transactionInfo == nil {
		l.svcCtx.UserOrder.Delete(l.ctx, order.Id)
		return &types.DeletOrderResp{Code: "10000", Msg: "yes"}, nil
	}
	if OrderCanBeDeleted(order, transactionInfo) {
		l.svcCtx.UserOrder.Delete(l.ctx, order.Id)
		l.svcCtx.TransactionInfo.Delete(l.ctx, transactionInfo.Id)
		return &types.DeletOrderResp{Code: "10000", Msg: "yes"}, nil
	}
	return &types.DeletOrderResp{Code: "10000", Msg: "订单状态不可删除"}, nil
}
