package coupon

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetvoucherLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetvoucherLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetvoucherLogic {
	return &GetvoucherLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetvoucherLogic) Getvoucher(req *types.GetVoucherRes) (resp *types.GetVoucherResp, err error) {
	// todo: add your logic here and delete this line

	return
}
