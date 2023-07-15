package orderpay

import (
	"context"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishordersLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	usecoupon bool
	usecash   bool
	ul        *UtilLogic
	userphone string
	oul       *OrderUtilLogic
}

func NewFinishordersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishordersLogic {
	return &FinishordersLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		ul:        NewUtilLogic(ctx, svcCtx),
		userphone: ctx.Value("phone").(string),
		oul:       NewOrderUtilLogic(ctx, svcCtx),
	}
}

func (l *FinishordersLogic) Finishorders(req *types.FinishOrdersRes) (resp *types.FinishOrdersResp, err error) {
	// 准备阶段，默认不用优惠券，不用钱包
	OrderInfos := make([]*types.OrderInfo, 0)
	l.usecoupon = false
	l.usecash = false
	orders, _ := l.svcCtx.Order.FindAllByOutTradeNo(l.ctx, req.OutTradeNo)
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, req.OutTradeNo)
	if orders == nil || len(orders) <= 0 || PayInfo == nil {
		return &types.FinishOrdersResp{Code: "4004", Msg: "数据库失效，请重新下单"}, nil
	}
	if PayInfo.CashAccountPayAmount > 0 {
		l.usecash = true
	}
	for _, order := range orders {
		if len(order.UsedCouponinfo) > 4 {
			l.usecoupon = true
		}
	}
	if l.usecash || l.usecoupon {
		lockmsglist := make([]*types.LockMsg, 0)
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.userphone, Field: "user_coupon"})
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.userphone, Field: "cash_account"})
		if l.ul.Getlocktry(lockmsglist) {
			cashok := false
			couponok := false
			okstr := ""
			if l.usecash { // 现金账户部分
				cashok, okstr = l.oul.Updatecashaccount(PayInfo.Phone, "OrderSn", PayInfo.OutTradeNo, PayInfo.CashAccountPayAmount, PayInfo.LogId, true)
				if !cashok || okstr != "yes" {
					l.ul.Oplog("支付模块更新现金账户失败", PayInfo.OutTradeNo, "开始更新", PayInfo.LogId)
				}
			} else {
				cashok = true
			}
			if l.usecoupon {
				couponok, okstr = l.FinishCoupon(orders, true)
				if !couponok || okstr != "yes" {
					l.ul.Oplog("支付模块更新优惠券失败", PayInfo.OutTradeNo, "开始更新", PayInfo.LogId)
				}
			} else {
				couponok = true
			}
			l.ul.Closelock(lockmsglist)
			if cashok && couponok {
				l.svcCtx.Order.UpdateCashPay(l.ctx, 1, PayInfo.OutTradeNo)
			}
			orders, _ = l.svcCtx.Order.FindAllByOutTradeNo(l.ctx, PayInfo.OutTradeNo)
			for _, order := range orders {
				OrderInfos = append(OrderInfos, OrderDb2info(order))
			}
			return &types.FinishOrdersResp{Code: "10000", Msg: "完全成功", Data: OrderInfos}, nil
		} else {
			return &types.FinishOrdersResp{Code: "10000", Msg: "未获取到锁，请重试", Data: OrderInfos}, nil
		}

	}
	return &types.FinishOrdersResp{Code: "10000", Msg: "不需要操作钱包或优惠券或积分，直接返回"}, nil
}
func (l FinishordersLogic) FinishCoupon(orders []*cachemodel.Order, use bool) (bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	for _, order := range orders {
		if len(order.UsedCouponinfo) > 4 {
			couponok, okstr := l.oul.UpdateCoupon(order, use)
			if !couponok || okstr != "yes" {
				l.ul.Oplog("支付模块更新优惠券失败", order.OrderSn, "开始更新", order.LogId)
				return false, "no"
			}
		}
	}
	return true, "yes"
}
