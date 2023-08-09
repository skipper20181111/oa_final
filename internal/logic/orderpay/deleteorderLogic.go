package orderpay

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteorderLogic {
	return &DeleteorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteorderLogic) Deleteorder(req *types.DeletOrderRes) (resp *types.DeletOrderResp, err error) {
	resp = &types.DeletOrderResp{Code: "10000", Msg: "删除成功"}
	OutTradeSnMap := make(map[string]bool)
	for _, OrderSn := range req.OrderSn {
		ok, OutTradeSn := l.DeletOneOrder(OrderSn)
		OutTradeSnMap[OutTradeSn] = true
		if !ok {
			resp = &types.DeletOrderResp{Code: "4004", Msg: "订单中部分商品暂无法删除，请您核实后再进行尝试"}
		}
	}
	for OutTradeSn, _ := range OutTradeSnMap {
		deleted, _ := l.svcCtx.Order.FindAllByOutTradeNoNotDeleted(l.ctx, OutTradeSn)
		if deleted != nil && len(deleted) == 0 {
			l.svcCtx.PayInfo.UpdateStatus(l.ctx, OutTradeSn, 9)
		}
	}
	return resp, nil
}
func (l DeleteorderLogic) DeletOneOrder(OrderSn string) (bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	order, _ := l.svcCtx.Order.FindOneByOrderSn(l.ctx, OrderSn)
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)
	if PayInfo == nil {
		l.svcCtx.Order.Delete(l.ctx, order.Id)
		return true, ""
	}
	if OrderCanBeDeleted(order, PayInfo) {
		l.svcCtx.Order.UpdateStatusByOrderSn(l.ctx, 9, order.OrderSn)
		return true, PayInfo.OutTradeNo
	}
	return false, PayInfo.OutTradeNo
}
