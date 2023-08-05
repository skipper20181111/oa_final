package deliver

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GivesfLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGivesfLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GivesfLogic {
	return &GivesfLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GivesfLogic) Givesf(req *types.GiveSFRes) (resp *types.GiveSFResp, err error) {
	// todo: add your logic here and delete this line

	return
}
