package refresh

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/internal/svc"
)

type ProbeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProbeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProbeLogic {
	return &ProbeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProbeLogic) Probe() error {

	return nil
}
