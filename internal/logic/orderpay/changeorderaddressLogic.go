package orderpay

import (
	"context"
	"encoding/json"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeorderaddressLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
	sf        *SfUtilLogic
	u         *UtilLogic
}

func NewChangeorderaddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeorderaddressLogic {
	return &ChangeorderaddressLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		userphone: ctx.Value("phone").(string),
		sf:        NewSfUtilLogic(ctx, svcCtx),
		u:         NewUtilLogic(ctx, svcCtx),
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
	Addressbytes, _ := json.Marshal(Address)
	if Getsha512(string(Addressbytes)) == Getsha512(sn2order.Address) || !strings.Contains(sn2order.Address, Address.Province) {
		return l.u.OrderDb2info(sn2order), true
	}
	go l.sf.RefundSfOrder(*sn2order)
	sn2order.Address = string(Addressbytes)
	sn2order.DeliverySn = ""
	sn2order.OrderStatus = 1
	err = l.svcCtx.Order.UpdateAddress(l.ctx, OrderSn, string(Addressbytes))
	if err != nil {
		return nil, false
	}
	return l.u.OrderDb2info(sn2order), true
}
