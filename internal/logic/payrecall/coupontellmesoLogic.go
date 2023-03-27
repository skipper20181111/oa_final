package payrecall

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CoupontellmesoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCoupontellmesoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CoupontellmesoLogic {
	return &CoupontellmesoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CoupontellmesoLogic) Coupontellmeso(req *types.TellMeSoRes) (resp *types.TellMeSoResp, err error) {
	// todo: add your logic here and delete this line

	return
}
