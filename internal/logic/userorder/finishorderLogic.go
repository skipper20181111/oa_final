package userorder

import (
	"context"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishorderLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	usecash     bool
	usecoupon   bool
	usepoint    bool
	cashaccount *cachemodel.CashAccount
	userorder   *cachemodel.UserOrder
}

func NewFinishorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishorderLogic {
	return &FinishorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FinishorderLogic) Finishorder(req *types.FinishOrderRes) (resp *types.FinishOrderResp, err error) {
	// 准备阶段，默认不用优惠券，不用钱包，不用积分
	l.usecoupon = false
	l.usecash = false
	l.usepoint = false
	lu := NewLogic(l.ctx, l.svcCtx)
	userphone := l.ctx.Value("phone").(string)
	userpoint := &cachemodel.UserPoints{}
	// 根据ordersn获取order信息 判定究竟使用什么，这三个应当是独立的，不应写在一起
	order, _ := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	transactioninfo, _ := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if order == nil || transactioninfo == nil {
		return &types.FinishOrderResp{Code: "4004", Msg: "数据库失效，请重新下单"}, nil
	}
	l.userorder = order
	if transactioninfo.CashAccountPayAmount > 0 {
		l.usecash = true
	}
	if l.userorder.UsedCouponinfo != "" {
		l.usecoupon = true
	}
	if l.userorder.PointAmount > 0 {
		l.usepoint = true
		userpoint, _ = l.svcCtx.UserPoints.FindOneByPhoneNoCache(l.ctx, userphone)
		if userpoint != nil {
			userpoint.AvailablePoints = userpoint.AvailablePoints - l.userorder.PointAmount
			l.svcCtx.UserPoints.Update(l.ctx, userpoint)
		}
	}
	// 第三阶段，挨个更新，如果更新失败，要回滚的。而且要告诉前端支付失败。同时要更新order界面，更新失败。那么order日志上，是否也要更新失败呢？
	if l.usecash || l.usecoupon || l.usepoint {
		lockmsglist := make([]*types.LockMsg, 0)
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: userphone, Field: "user_coupon"})
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: userphone, Field: "cash_account"})
		if lu.getlock(lockmsglist) {
			if l.usecash { // 现金账户部分
				cashok, okstr := lu.Updatecashaccount(l.userorder, true)
				if !cashok || okstr != "yes" {
					lu.Oplog("支付模块更新现金账户失败", order.OrderSn, "开始更新", l.userorder.LogId)
				} else {
					userpoint.HistoryPoints = userpoint.HistoryPoints + l.userorder.CashAccountPayAmount
					userpoint.AvailablePoints = userpoint.AvailablePoints + l.userorder.CashAccountPayAmount
					l.svcCtx.UserPoints.Update(l.ctx, userpoint)
				}
			}
			if l.usecoupon {
				couponok, okstr := lu.UpdateCoupon(l.userorder, true)
				if !couponok || okstr != "yes" {
					lu.Oplog("支付模块更新优惠券失败", order.OrderSn, "开始更新", l.userorder.LogId)
				}
			}
			lu.closelock(lockmsglist)
			l.svcCtx.UserOrder.FinishAccountPay(l.ctx, order.OrderSn)
			return &types.FinishOrderResp{Code: "10000", Msg: "完全成功", Data: OrderDb2info(order, transactioninfo)}, nil
		} else {
			return &types.FinishOrderResp{Code: "10000", Msg: "未获取到锁，请重试", Data: OrderDb2info(order, transactioninfo)}, nil
		}
	}

	return &types.FinishOrderResp{Code: "10000", Msg: "不需要操作钱包或优惠券或积分，直接返回"}, nil
}
