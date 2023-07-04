package payrecall

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/cachemodel"
	"oa_final/internal/logic/orderpay"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"time"
)

type CoupontellmesoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	lu     *orderpay.UtilLogic
}

func NewCoupontellmesoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CoupontellmesoLogic {

	return &CoupontellmesoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		lu:     orderpay.NewUtilLogic(ctx, svcCtx),
	}
}

func (l *CoupontellmesoLogic) Coupontellmeso(notifyReq *notify.Request, transaction *payments.Transaction) (resp *types.TellMeSoResp, err error) {
	fmt.Println("************** START ******************")
	if *transaction.TradeState == "SUCCESS" {
		order, _ := l.svcCtx.RechargeOrder.FindOneByOutTradeNo(l.ctx, *transaction.OutTradeNo)
		if order != nil {
			l.ctx = context.WithValue(l.ctx, "phone", "17854230845")
			l.ctx = context.WithValue(l.ctx, "openid", "17854230845")
			lockmsglist := make([]*types.LockMsg, 0)
			lockmsglist = append(lockmsglist, &types.LockMsg{Phone: order.OutTradeNo, Field: "user_coupon"})
			if l.lu.Getlocktry(lockmsglist) {
				l.lu.Oplog("现金账户充值", order.OrderSn, "开始更新", order.LogId)
				phone, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, order.Phone)
				totleamount := order.Amount + order.GiftAmount
				if phone == nil {
					_, err := l.svcCtx.CashAccount.Insert(l.ctx, &cachemodel.CashAccount{Phone: order.Phone, Balance: totleamount})
					if err != nil {
						l.lu.Closelock(lockmsglist)
						return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
					}
					l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), OrderType: "充值", OrderSn: order.OrderSn, OrderDescribe: "微信支付充值送现金", Behavior: "充值", Phone: order.Phone, Balance: totleamount, ChangeAmount: totleamount})
					l.lu.Closelock(lockmsglist)
				} else {
					phone.Balance = phone.Balance + totleamount
					l.svcCtx.CashAccount.Update(l.ctx, phone)
					l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), OrderType: "充值", OrderSn: order.OrderSn, OrderDescribe: "微信支付充值送现金", Behavior: "充值", Phone: order.Phone, Balance: phone.Balance, ChangeAmount: totleamount})
					l.lu.Closelock(lockmsglist)
				}
				l.svcCtx.RechargeOrder.UpdateFinished(l.ctx, order.OutTradeNo, *transaction.TransactionId, time.Now())
				l.lu.Oplog("现金账户充值", order.OrderSn, "结束更新", order.LogId)
			} else {
				return &types.TellMeSoResp{Code: "FAIL", Message: "未获取到锁"}, nil
			}
			fmt.Println("*************** END *******************")
		}

	}

	return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
}
