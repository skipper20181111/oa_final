package payrecall

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"oa_final/internal/logic/userorder"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type TellmesoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTellmesoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TellmesoLogic {
	return &TellmesoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TellmesoLogic) Tellmeso(notifyReq *notify.Request, transaction *payments.Transaction) (resp *types.TellMeSoResp, err error) {
	fmt.Println("************** START ******************")
	lu := userorder.NewLogic(l.ctx, l.svcCtx)

	if *transaction.TradeState == "SUCCESS" {

		no, _ := l.svcCtx.UserOrder.FindOneByOutTradeNo(l.ctx, *transaction.OutTradeNo)
		l.svcCtx.UserOrder.FinishAccountPay(l.ctx, no.OrderSn)
		lu.Oplog("user_order", *transaction.OutTradeNo, "开始更新", no.LogId)
		lu.Oplog("付款啊", no.OrderSn, "结束更新", no.LogId)
		if no != nil {
			cache, _ := l.svcCtx.UserPoints.FindOneByPhoneNoCache(l.ctx, no.Phone)
			if cache != nil {
				cache.HistoryPoints = cache.HistoryPoints + no.WexinPayAmount
				cache.AvailablePoints = cache.AvailablePoints + no.WexinPayAmount
				l.svcCtx.UserPoints.Update(l.ctx, cache)
			}
			if no.OrderStatus == 0 {
				no.TransactionId = *transaction.TransactionId
				no.PaymentTime = time.Now()
				no.ModifyTime = time.Now()
				l.svcCtx.UserOrder.Update(l.ctx, no)
				lu.Oplog("user_order", *transaction.OutTradeNo, "结束更新", no.LogId)
				return &types.TellMeSoResp{Code: "SUCCESS", Message: "成功"}, nil
			} else {
				no.TransactionId = *transaction.TransactionId
				no.PaymentTime = time.Now()
				no.ModifyTime = time.Now()
				l.svcCtx.UserOrder.Update(l.ctx, no)
				return &types.TellMeSoResp{Code: "SUCCESS", Message: "成功"}, nil
			}
		}
	}
	fmt.Println("*************** END *******************")
	return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
}
