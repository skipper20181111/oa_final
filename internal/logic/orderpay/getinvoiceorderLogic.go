package orderpay

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetinvoiceorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	phone  string
}

func NewGetinvoiceorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetinvoiceorderLogic {
	return &GetinvoiceorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		phone:  ctx.Value("phone").(string),
	}
}

func (l *GetinvoiceorderLogic) Getinvoiceorder(req *types.GetInvoiceOrderRes) (resp *types.GetAllOrderResp, err error) {
	userphone := l.ctx.Value("phone").(string)
	resp = &types.GetAllOrderResp{
		Code: "10000",
		Msg:  "success",
		Data: &types.GetAllOrderRp{
			OrderInfos: make([]*types.OrderInfo, 0),
		},
	}
	payInfos, _ := l.svcCtx.PayInfo.FindStatus4(l.ctx, req.PageNumber)
	Orders, _ := l.svcCtx.Order.FindAllByOutTradeNos(l.ctx, userphone, payInfos)
	for _, order := range Orders {
		resp.Data.OrderInfos = append(resp.Data.OrderInfos, OrderDb2info(order))
	}
	return resp, nil
}
