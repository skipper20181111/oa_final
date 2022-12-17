package userorder

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetallorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetallorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetallorderLogic {
	return &GetallorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetallorderLogic) Getallorder(req *types.GetAllOrderRes) (resp *types.GetAllOrderResp, err error) {
	// todo: add your logic here and delete this line

	return
}
