package refresh

import (
	"context"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshPLLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshPLLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshPLLogic {
	return &RefreshPLLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshPLLogic) RefreshPL() (resp *types.RefreshResp, err error) {
	refreshUtil := NewRefreshUtilLogic(l.ctx, l.svcCtx)
	refreshUtil.InfoMapAndMap()
	refreshUtil.RechargeProduct()
	refreshUtil.StarMall()
	refreshUtil.Coupon()
	refreshUtil.MissionList()
	return &types.RefreshResp{Code: "10000", Msg: "刷新成功"}, err
}
