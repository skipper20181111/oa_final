package orderpay

import (
	"context"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmorderLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
}

func NewConfirmorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmorderLogic {
	return &ConfirmorderLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		userphone: ctx.Value("phone").(string),
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
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, req.OutTradeNo)
	if PayInfo == nil || PayInfo.Status == 4 {
		return resp, nil
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
		l.svcCtx.UserPoints.UpdatePoints(l.ctx, PayInfo.Phone, PayInfo.TotleAmount)
		l.svcCtx.Order.UpdateClosedByOutTradeSn(l.ctx, req.OutTradeNo)
		userPoints, _ := l.svcCtx.UserPoints.FindOneByPhone(l.ctx, l.userphone)
		l.svcCtx.PointLog.Insert(l.ctx, &cachemodel.PointLog{Date: time.Now(),
			OrderType:     "正常商品",
			OrderSn:       PayInfo.OutTradeNo,
			OrderDescribe: "正常商品收货获取积分",
			Behavior:      "获取",
			Phone:         l.userphone,
			Balance:       userPoints.AvailablePoints,
			ChangeAmount:  PayInfo.TotleAmount/100 + 1,
		})

	}
	orders, _ := l.svcCtx.Order.FindAllByOutTradeNo(l.ctx, req.OutTradeNo)
	for _, order := range orders {
		resp.Data.OrderInfos = append(resp.Data.OrderInfos, OrderDb2info(order))
	}
	return resp, nil
}
