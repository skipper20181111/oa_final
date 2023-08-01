package orderpay

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	col    *CheckOrderLogic
}

func NewGetorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetorderLogic {
	return &GetorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		col:    NewCheckOrderLogic(ctx, svcCtx),
	}
}

func (l *GetorderLogic) Getorder(req *types.GetOrderRes) (resp *types.GetAllOrderResp, err error) {
	// todo: add your logic here and delete this line

	return l.col.MonitorOrderStatus(req.OutTradeNo)
}
