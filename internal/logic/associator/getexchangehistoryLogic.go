package associator

import (
	"context"
	"oa_final/internal/logic/orderpay"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetexchangehistoryLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
	u         *orderpay.UtilLogic
}

func NewGetexchangehistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetexchangehistoryLogic {
	return &GetexchangehistoryLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		userphone: ctx.Value("phone").(string),
		u:         orderpay.NewUtilLogic(ctx, svcCtx),
	}
}

func (l *GetexchangehistoryLogic) Getexchangehistory(req *types.GetExchangeHistoryRes) (resp *types.GetExchangeHistoryResp, err error) {
	resp = &types.GetExchangeHistoryResp{
		Code: "10000",
		Msg:  "success",
		Data: &types.GetExchangeHistoryRp{
			OrderList: make([]*types.OrderInfo, 0),
		},
	}
	orders, _ := l.svcCtx.Order.FindAllPointsCouponOrder(l.ctx, l.userphone)
	for _, order := range orders {
		resp.Data.OrderList = append(resp.Data.OrderList, l.u.OrderDb2info(order))
	}
	return resp, nil
}
