package deliver

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetordersnLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetordersnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetordersnLogic {
	return &GetordersnLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetordersnLogic) Getordersn(req *types.GetOrderSnRes) (resp *types.GetOrderSnResp, err error) {

	return
}
