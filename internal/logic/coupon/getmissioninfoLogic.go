package coupon

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetmissioninfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	mu     *MissionUtilLogic
}

func NewGetmissioninfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetmissioninfoLogic {
	return &GetmissioninfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		mu:     NewMissionUtilLogic(ctx, svcCtx),
	}
}

func (l *GetmissioninfoLogic) Getmissioninfo(req *types.GetMissionInfoRes) (resp *types.GetMissionInfoResp, err error) {
	return l.mu.Getmissioninfo()
}
