package payrecall

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zeromicro/go-zero/core/logx"
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
		sn, _ := l.svcCtx.UserOrder.FindOneByOutTradeNo(l.ctx, *transaction.OutTradeNo)
		if sn != nil {
			l.svcCtx.TransactionInfo.UpdateWeixinPay(l.ctx, sn.OrderSn, *transaction.TransactionId)
			sn.TransactionId = *transaction.TransactionId
			l.svcCtx.UserOrder.Update(l.ctx, sn)
			fmt.Println("*************** END *******************")
			return &types.TellMeSoResp{Code: "SUCCESS", Message: "成功"}, nil
		} else {
			return &types.TellMeSoResp{Code: "SUCCESS", Message: "成功"}, nil
		}
	}

	return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
}
