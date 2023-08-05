package deliver

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreparegoodsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreparegoodsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreparegoodsLogic {
	return &PreparegoodsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreparegoodsLogic) Preparegoods(req *types.PrepareGoodsRes) (resp *types.PrepareGoodsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
