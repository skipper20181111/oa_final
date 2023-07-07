package orderpay

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
	col    *CheckOrderLogic
}

func NewGetallorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetallorderLogic {
	return &GetallorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		col:    NewCheckOrderLogic(ctx, svcCtx),
	}
}

func (l *GetallorderLogic) Getallorder(req *types.GetAllOrderRes) (resp *types.GetAllOrderResp, err error) {
	infos := make([]*types.OrderInfo, 0)
	userphone := l.ctx.Value("phone").(string)
	allorder, err := l.svcCtx.Order.FindAllByPhone(l.ctx, userphone, req.PageNumber)
	if allorder == nil || len(allorder) == 0 {
		return &types.GetAllOrderResp{Code: "10000", Msg: "success", Data: &types.GetAllOrderRp{OrderInfos: infos}}, nil
	}
	for _, order := range allorder {
		if OrderNeedChange(order) {
			sn, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)
			order = l.col.checkall(order, sn)
		}
		infos = append(infos, OrderDb2info(order))
	}
	return &types.GetAllOrderResp{Code: "10000", Msg: "success", Data: &types.GetAllOrderRp{OrderInfos: infos}}, nil
}
