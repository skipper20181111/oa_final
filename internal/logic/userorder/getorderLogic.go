package userorder

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

type GetorderLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	checklogic *CheckOrderLogic
}

func NewGetorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetorderLogic {
	return &GetorderLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		checklogic: NewCheckOrderLogic(ctx, svcCtx),
	}
}
func (l *GetorderLogic) Getorder(req *types.GetOrderRes) (resp *types.GetOrderResp, err error) {
	return l.checklogic.MonitorOrderStatus(req.OrderSn)
}
