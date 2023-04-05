package userorder

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PrepayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPrepayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PrepayLogic {
	return &PrepayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PrepayLogic) Prepay(req *types.FinishOrderRes) (resp *types.PrePayResp, err error) {
	order, err := l.svcCtx.UserOrdercache.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil {
		return &types.PrePayResp{Code: "10000", Msg: "未查询到订单,请重建订单", Data: false}, nil
	}

	cashaccount, _ := l.svcCtx.CashAccount.FindOneByPhoneNoCach(l.ctx, order.Phone)
	if cashaccount == nil || cashaccount.Balance < order.CashAccountPayAmount {
		return &types.PrePayResp{Code: "10000", Msg: "账户余额不足，请重建订单", Data: false}, nil
	}
	usercoupon, _ := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, order.Phone)
	if usercoupon == nil {
		return &types.PrePayResp{Code: "10000", Msg: "未查询到优惠券，请重建订单", Data: false}, nil
	} else {
		if !HaveKey(order.UsedCouponinfo, usercoupon.CouponIdMap) {
			return &types.PrePayResp{Code: "10000", Msg: "未查询到优惠券，请重建订单", Data: false}, nil
		}
	}

	return &types.PrePayResp{Code: "10000", Msg: "此订单可支付", Data: true}, nil
}
