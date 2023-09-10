package orderpay

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GettransactionidLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGettransactionidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GettransactionidLogic {
	return &GettransactionidLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GettransactionidLogic) Gettransactionid(req *types.GetTransactionIDRes) (resp *types.GetTransactionIDResp, err error) {
	resp = &types.GetTransactionIDResp{
		Code: "10000",
		Msg:  "success",
		Data: &types.GetTransactionIDRp{
			TransactionId: "",
		},
	}
	payInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, req.OutTradeNo)
	resp.Data.TransactionId = payInfo.TransactionId
	return resp, nil
}
