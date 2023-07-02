package coupon

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishmissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	mu     *MissionUtilLogic
}

func NewFinishmissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishmissionLogic {
	return &FinishmissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		mu:     NewMissionUtilLogic(ctx, svcCtx),
	}
}

func (l *FinishmissionLogic) Finishmission(req *types.FinishMissionRes) (resp *types.GetMissionInfoResp, err error) {
	return l.mu.Finishmission(req.MissionId)
}
