package payrecall

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/internal/logic/orderpay"
	"oa_final/internal/svc"
	"oa_final/internal/types"
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
	if *transaction.TradeState == "SUCCESS" {
		l.svcCtx.PayInfo.UpdateWeixinPay(l.ctx, *transaction.OutTradeNo, *transaction.TransactionId)
		PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, *transaction.OutTradeNo)
		if PayInfo != nil && PayInfo.FinishWeixinpay == 1 {
			if total, _, _ := orderpay.IfFinished(PayInfo); total {
				l.svcCtx.PayInfo.UpdateAllPay(l.ctx, *transaction.OutTradeNo)
				l.svcCtx.Order.UpdateStatusByOutTradeSn(l.ctx, 1, *transaction.OutTradeNo)
			}
			fmt.Println("*************** END *******************")
			return &types.TellMeSoResp{Code: "SUCCESS", Message: "成功"}, nil
		}
	}
	return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
}
