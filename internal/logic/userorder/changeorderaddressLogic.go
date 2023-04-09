package userorder

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeorderaddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangeorderaddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeorderaddressLogic {
	return &ChangeorderaddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangeorderaddressLogic) Changeorderaddress(req *types.ChangeOrdeRaddressRes) (resp *types.ChangeOrdeRaddressResp, err error) {
	userphone := l.ctx.Value("phone").(string)
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	if sn2order == nil {
		fmt.Println(err.Error())
	}
	if sn2order.Phone != userphone {
		return &types.ChangeOrdeRaddressResp{
			Code: "4004",
			Msg:  "请不要使用别人的token",
		}, nil
	}
	addr, err := json.Marshal(req.Address)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	sn2order.Address = string(addr)
	sn2order.ModifyTime = time.Now()
	err = l.svcCtx.UserOrder.Update(l.ctx, sn2order)
	if err != nil {
		fmt.Println(err.Error())
	}
	sn, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	info, err := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if sn == nil || sn.Address != string(addr) {
		return &types.ChangeOrdeRaddressResp{
			Code: "4004",
			Msg:  "数据库失效",
		}, nil
	}
	return &types.ChangeOrdeRaddressResp{Code: "10000", Msg: "修改成功", Data: &types.ChangeOrdeRaddressRp{OrderInfo: OrderDb2info(sn, info)}}, nil
}
