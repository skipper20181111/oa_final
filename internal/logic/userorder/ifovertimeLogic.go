package userorder

import (
	"context"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IfovertimeLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	CheckOrderLogic *CheckOrderLogic
}

func NewIfovertimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IfovertimeLogic {
	return &IfovertimeLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		CheckOrderLogic: NewCheckOrderLogic(ctx, svcCtx),
	}
}

func (l *IfovertimeLogic) Ifovertime(req *types.IfOvertimeRes) (resp *types.IfOvertimeResp, err error) {
	order, _ := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	transactionInfo, _ := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil {
		return
	}
	if transactionInfo == nil {
		return
	}
	if OrderCanBeDeleted(order, transactionInfo) {
		order.OrderStatus = 8
		l.svcCtx.UserOrder.UpdateByOrderSn(l.ctx, order)
		return
	}
	return
}
