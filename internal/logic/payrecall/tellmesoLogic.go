package payrecall

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"math/rand"
	"oa_final/cachemodel"
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
	lid := time.Now().UnixNano() - int64(rand.Intn(1024))
	l.oplog("user_order", *transaction.OutTradeNo, "开始更新", lid)
	if *transaction.TradeState == "SUCCESS" {
		no, _ := l.svcCtx.UserOrder.FindOneByOutTradeNo(l.ctx, *transaction.OutTradeNo)
		if no != nil {
			if no.OrderStatus == 0 {
				no.OrderStatus = 1
				no.TransactionId = *transaction.TransactionId
				l.svcCtx.UserOrder.Update(l.ctx, no)
				l.oplog("user_order", *transaction.OutTradeNo, "结束更新", lid)
				return &types.TellMeSoResp{Code: "SUCCESS", Message: "成功"}, nil
			} else {
				no.OrderStatus = 99
				no.TransactionId = *transaction.TransactionId
				l.svcCtx.UserOrder.Update(l.ctx, no)
				return &types.TellMeSoResp{Code: "SUCCESS", Message: "成功"}, nil
			}
		}
	}
	fmt.Println("*************** END *******************")
	return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
}
func (l *TellmesoLogic) oplog(tablename, event, describe string, lid int64) error {
	aol := &cachemodel.AccountOperateLog{Phone: l.ctx.Value("phone").(string), TableName: tablename, Event: event, Describe: describe, Timestamp: time.Now(), Lid: lid}
	_, err := l.svcCtx.AccountOperateLog.Insert(l.ctx, aol)
	return err
}
