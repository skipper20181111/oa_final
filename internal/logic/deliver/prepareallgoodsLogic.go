package deliver

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PrepareallgoodsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPrepareallgoodsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PrepareallgoodsLogic {
	return &PrepareallgoodsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PrepareallgoodsLogic) Prepareallgoods(req *types.PrepareAllGoodsRes) (resp *types.GiveSFResp, err error) {
	l.svcCtx.Order.PrepareAllGoods(l.ctx, req.MarketID)
	return &types.GiveSFResp{
		Code: "10000",
		Msg:  "success",
	}, nil
}
