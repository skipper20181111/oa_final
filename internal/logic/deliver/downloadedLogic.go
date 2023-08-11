package deliver

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadedLogic {
	return &DownloadedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadedLogic) Downloaded(req *types.DownLoadedRes) (resp *types.DownLoadedResp, err error) {
	resp = &types.DownLoadedResp{
		Sfsn1002: make([]string, 0),
	}
	for _, sfsn := range req.SfSn {
		l.svcCtx.Order.FinishDownload(l.ctx, 1002, sfsn)
	}
	orders, _ := l.svcCtx.Order.FindAll1002(l.ctx)
	for _, order := range orders {
		resp.Sfsn1002 = append(resp.Sfsn1002, order.DeliverySn)
	}
	return resp, nil
}
