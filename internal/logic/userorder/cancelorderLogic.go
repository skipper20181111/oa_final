package userorder

import (
	"context"

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
	// todo: add your logic here and delete this line

	return
}
