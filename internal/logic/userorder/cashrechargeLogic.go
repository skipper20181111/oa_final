package userorder

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CashrechargeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCashrechargeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CashrechargeLogic {
	return &CashrechargeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CashrechargeLogic) Cashrecharge(req *types.CashRechargeRes) (resp *types.CashRechargeResp, err error) {

	return
}
