package deliver

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadsfpdfLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadsfpdfLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadsfpdfLogic {
	return &DownloadsfpdfLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadsfpdfLogic) Downloadsfpdf() (resp *types.DownloadSFPDFResp, err error) {
	resp = &types.DownloadSFPDFResp{
		SfPDFs: make([]string, 0),
	}
	orders, _ := l.svcCtx.Order.FindAll1001(l.ctx)
	for _, order := range orders {
		resp.SfPDFs = append(resp.SfPDFs, order.DeliverySn)
	}
	return resp, nil
}
