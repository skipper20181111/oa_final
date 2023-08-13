package orderpay

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmorderLogic {
	return &ConfirmorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmorderLogic) Confirmorder(req *types.ConfirmOrderRes) (resp *types.ConfirmOrderResp, err error) {
	resp = &types.ConfirmOrderResp{
		Code: "10000",
		Msg:  "success",
		Data: &types.GetAllOrderRp{
			OrderInfos: make([]*types.OrderInfo, 0),
		},
	}
	status, _ := l.svcCtx.Order.FindAllStatusByOutTradeNo(l.ctx, req.OutTradeNo)
	yes := true
	for _, sta := range status {
		if sta != 2 && sta != 3 && sta != 4 {
			yes = false
		}
	}
	if yes {
		l.svcCtx.PayInfo.UpdateStatus(l.ctx, req.OutTradeNo, 4)
		l.svcCtx.Order.UpdateStatusByOutTradeSn(l.ctx, 4, req.OutTradeNo)
	}
	orders, _ := l.svcCtx.Order.FindAllByOutTradeNo(l.ctx, req.OutTradeNo)
	for _, order := range orders {
		resp.Data.OrderInfos = append(resp.Data.OrderInfos, OrderDb2info(order))
	}
	return resp, nil
}
