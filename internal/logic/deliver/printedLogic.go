package deliver

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PrintedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPrintedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PrintedLogic {
	return &PrintedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PrintedLogic) Printed(req *types.DownLoadedRes) (resp *types.DownLoadedResp, err error) {
	// todo: add your logic here and delete this line

	return
}
