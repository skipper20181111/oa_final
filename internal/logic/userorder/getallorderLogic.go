package userorder

import (
	"context"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetallorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetallorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetallorderLogic {
	return &GetallorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetallorderLogic) Getallorder(req *types.GetAllOrderRes) (resp *types.GetAllOrderResp, err error) {
	userphone := l.ctx.Value("phone").(string)
	allorder, err := l.svcCtx.UserOrder.FindAllByPhone(l.ctx, userphone)
	if allorder == nil || len(allorder) == 0 {
		infos := make([]*types.OrderInfo, 0)
		return &types.GetAllOrderResp{Code: "10000", Msg: "success", Data: &types.GetAllOrderRp{OrderInfos: infos}}, nil
	}
	infos := make([]*types.OrderInfo, 0)
	for _, order := range allorder {
		orderinfo := OrderDb2info(order, nil)
		infos = append(infos, orderinfo)
	}

	return &types.GetAllOrderResp{Code: "10000", Msg: "success", Data: &types.GetAllOrderRp{OrderInfos: infos}}, nil
}
