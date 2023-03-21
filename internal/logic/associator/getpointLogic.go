package associator

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetpointLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetpointLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetpointLogic {
	return &GetpointLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetpointLogic) Getpoint(req *types.GetPointRes) (resp *types.GetPointResp, err error) {
	userphone := l.ctx.Value("phone").(string)
	//useropenid := l.ctx.Value("openid").(string)
	userpoint, _ := l.svcCtx.UserPoints.FindOneByPhone(l.ctx, userphone)
	if userpoint != nil {
		return &types.GetPointResp{Code: "10000", Msg: "success", Data: &types.GetPointRp{HistoryPoints: userpoint.HistoryPoints, AvailablePoints: userpoint.AvailablePoints}}, nil
	}

	return &types.GetPointResp{Code: "10000", Msg: "success", Data: &types.GetPointRp{HistoryPoints: 0, AvailablePoints: 0}}, nil

}
