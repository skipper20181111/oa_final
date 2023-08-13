package orderpay

import (
	"context"
	"encoding/json"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeorderaddressLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
}

func NewChangeorderaddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeorderaddressLogic {
	return &ChangeorderaddressLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		userphone: ctx.Value("phone").(string),
	}
}

func (l *ChangeorderaddressLogic) Changeorderaddress(req *types.ChangeOrdeRaddressRes) (resp *types.ChangeOrdeRaddressResp, err error) {
	resp = &types.ChangeOrdeRaddressResp{
		Code: "10000",
		Msg:  "success",
		Data: &types.ChangeOrdeRaddressRp{
			OrderInfoList: make([]*types.OrderInfo, 0),
		},
	}
	for _, OrderSn := range req.OrderSnList {
		Order, ok := l.ChangeAddress(OrderSn, req.Address)
		if ok {
			resp.Data.OrderInfoList = append(resp.Data.OrderInfoList, Order)
		}
	}
	return resp, nil
}
func (l ChangeorderaddressLogic) ChangeAddress(OrderSn string, Address *types.AddressInfo) (*types.OrderInfo, bool) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	sn2order, err := l.svcCtx.Order.FindOneByOrderSn(l.ctx, OrderSn)
	if sn2order == nil || sn2order.Phone != l.userphone || (sn2order.OrderStatus != 0 && sn2order.OrderStatus != 1 && sn2order.OrderStatus != 1000) {
		return nil, false
	}
	addr, err := json.Marshal(Address)
	if err != nil {
		return nil, false
	}
	sn2order.Address = string(addr)
	sn2order.ModifyTime = time.Now()
	err = l.svcCtx.Order.Update(l.ctx, sn2order)
	if err != nil {
		return nil, false
	}
	return OrderDb2info(sn2order), true
}
