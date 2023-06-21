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
	vu     *VoucherUtileLogic
}

func NewGetvoucherLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetvoucherLogic {
	return &GetvoucherLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		vu:     NewVoucherUtileLogic(ctx, svcCtx),
	}
}

func (l *GetvoucherLogic) Getvoucher(req *types.GetVoucherRes) (resp *types.GetVoucherResp, err error) {
	_, msg := l.vu.voucherbind(req.VoucherCode)
	return &types.GetVoucherResp{Code: "10000", Msg: msg}, nil
}
