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
		PDFList: make([]string, 0),
	}
	for _, sfsn := range req.SfSn {
		l.svcCtx.Order.UpdateStatusByDeliverySn(l.ctx, 1002, 1001, sfsn)
	}
	orders, _ := l.svcCtx.Order.FindAll1002(l.ctx)
	for _, order := range orders {
		resp.PDFList = append(resp.PDFList, order.DeliverySn)
	}
	return resp, nil
}
