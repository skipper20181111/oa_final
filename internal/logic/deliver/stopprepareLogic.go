package deliver

import (
	"context"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StopprepareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStopprepareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StopprepareLogic {
	return &StopprepareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StopprepareLogic) Stopprepare() (resp *types.NormalResp, err error) {
	l.svcCtx.Configuration.Update(l.ctx, &cachemodel.Configuration{
		Id:       3,
		Config:   "false",
		Describe: time.Now().Format("2006-01-02 15:04:05"),
	})
	l.svcCtx.Order.PrepareAllGoods(l.ctx, 0)
	return &types.NormalResp{
		Code: "10000",
		Msg:  "success",
	}, nil
}
