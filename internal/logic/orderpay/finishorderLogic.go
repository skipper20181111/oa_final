package orderpay

import (
	"context"
	"oa_final/cachemodel"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishorderLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	usecoupon bool
	usecash   bool
	ul        *UtilLogic
	userphone string
	userorder *cachemodel.Order
	oul       *OrderUtilLogic
}

func NewFinishorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishorderLogic {
	return &FinishorderLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		ul:        NewUtilLogic(ctx, svcCtx),
		userphone: ctx.Value("phone").(string),
		oul:       NewOrderUtilLogic(ctx, svcCtx),
	}
}

func (l *FinishorderLogic) Finishorder(req *types.FinishOrderRes) (resp *types.FinishOrderResp, err error) {
	// 准备阶段，默认不用优惠券，不用钱包
	l.usecoupon = false
	l.usecash = false
	userphone := l.ctx.Value("phone").(string)
	// 根据ordersn获取order信息 判定究竟使用什么，这三个应当是独立的，不应写在一起
	order, _ := l.svcCtx.Order.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil {
		return &types.FinishOrderResp{Code: "4004", Msg: "数据库失效，请重新下单"}, nil
	}
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)
	if PayInfo == nil {
		return &types.FinishOrderResp{Code: "4004", Msg: "数据库失效，请重新下单"}, nil
	}
	if PayInfo.FinishAccountpay == 1 {
		return &types.FinishOrderResp{Code: "10000", Msg: "已经完成支付，请勿重复支付"}, nil
	}
	l.userorder = order
	if PayInfo.CashAccountPayAmount > 0 {
		l.usecash = true
	}
	if l.userorder.UsedCouponinfo != "" {
		l.usecoupon = true
	}
	// 第三阶段，挨个更新，如果更新失败，要回滚的。而且要告诉前端支付失败。同时要更新order界面，更新失败。那么order日志上，是否也要更新失败呢？
	if l.usecash || l.usecoupon {
		lockmsglist := make([]*types.LockMsg, 0)
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: userphone, Field: "user_coupon"})
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: userphone, Field: "cash_account"})
		if l.ul.Getlocktry(lockmsglist) {
			if l.usecash { // 现金账户部分
				cashok, okstr := l.oul.Updatecashaccount(l.userorder, true)
				if !cashok || okstr != "yes" {
					l.ul.Oplog("支付模块更新现金账户失败", order.OrderSn, "开始更新", l.userorder.LogId)
				}
			}
			if l.usecoupon {
				couponok, okstr := l.oul.UpdateCoupon(l.userorder, true)
				if !couponok || okstr != "yes" {
					l.ul.Oplog("支付模块更新优惠券失败", order.OrderSn, "开始更新", l.userorder.LogId)
				}
			}
			l.ul.Closelock(lockmsglist)
			l.userorder.FinishAccountpay = 1
			l.svcCtx.Order.Update(l.ctx, l.userorder)
			return &types.FinishOrderResp{Code: "10000", Msg: "完全成功", Data: OrderDb2info(order)}, nil
		} else {
			return &types.FinishOrderResp{Code: "10000", Msg: "未获取到锁，请重试", Data: OrderDb2info(order)}, nil
		}
	}

	return &types.FinishOrderResp{Code: "10000", Msg: "不需要操作钱包或优惠券或积分，直接返回"}, nil
}
