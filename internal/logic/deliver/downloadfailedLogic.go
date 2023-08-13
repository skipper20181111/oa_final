package deliver

import (
	"context"
	"oa_final/internal/logic/orderpay"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadfailedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadfailedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadfailedLogic {
	return &DownloadfailedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadfailedLogic) Downloadfailed(req *types.DownLoadedRes) error {
	l.GetSfSnPDF(req.SfSn[0])
	return nil
}
func (l DownloadfailedLogic) GetSfSnPDF(SfSn string) {
	sf := orderpay.NewSfUtilLogic(context.Background(), l.svcCtx)
	order, _ := l.svcCtx.Order.FindOneBySfSn(l.ctx, SfSn)
	sf.GetPDF(order, SfSn)
}
