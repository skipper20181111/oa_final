package payrecall

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"net/http"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CoupontellmesoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCoupontellmesoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CoupontellmesoLogic {
	return &CoupontellmesoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CoupontellmesoLogic) Coupontellmeso(notifyReq *notify.Request, transaction *payments.Transaction) (resp *types.TellMeSoResp, err error) {
	fmt.Println("************** START ******************")
	if *transaction.TradeState == "SUCCESS" {
		order, _ := l.svcCtx.RechargeOrder.FindOneByOutTradeNo(l.ctx, *transaction.OutTradeNo)
		if order != nil {
			lockmsglist := make([]*types.LockMsg, 0)
			lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.ctx.Value("phone").(string), Field: "user_coupon"})
			if l.getlock(lockmsglist) {
				l.oplog("cashaccount", order.OrderSn, "开始更新", order.LogId)
				phone, err := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, order.Phone)
				if phone == nil && err.Error() == "notfind" {
					_, err := l.svcCtx.CashAccount.Insert(l.ctx, &cachemodel.CashAccount{Phone: order.Phone, Balance: order.WexinPayAmount})
					if err != nil {
						l.closelock(lockmsglist)
						return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
					}
				} else if phone != nil {
					phone.Balance = phone.Balance + order.WexinPayAmount
					l.svcCtx.CashAccount.Update(l.ctx, phone)
				} else {
					l.closelock(lockmsglist)
					return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
				}

				l.oplog("cashaccount", order.OrderSn, "结束更新", order.LogId)
				l.oplog("现金账户充值", order.OrderSn, "结束更新", order.LogId)
				l.closelock(lockmsglist)
			} else {
				return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
			}
		}
	}
	fmt.Println("*************** END *******************")
	return &types.TellMeSoResp{Code: "FAIL", Message: "失败"}, nil
}

func (l *CoupontellmesoLogic) getlock(lockmsglist []*types.LockMsg) bool {
	//phone := l.ctx.Value("phone").(string)
	lockhost := l.svcCtx.Config.Lock.Host
	urlPath := fmt.Sprintf("%s%s%s", "http://", lockhost, "/pcc/getlock")

	res := types.GetLockRes{LockMsgList: lockmsglist}

	resp, err := httpc.Do(context.Background(), http.MethodPost, urlPath, res)
	if err != nil {

		fmt.Println(err)
	}
	if resp == nil || resp.Body == nil {
		return false
	}
	lockresult := &types.GetLockResp{Code: make(map[string]bool)}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, lockresult)
	defer resp.Body.Close()
	for _, b := range lockresult.Code {
		if b == false {
			return false
		}
	}
	return true

}
func (l *CoupontellmesoLogic) closelock(lockmsglist []*types.LockMsg) bool {
	//phone := l.ctx.Value("phone").(string)
	lockhost := l.svcCtx.Config.Lock.Host
	urlPath := fmt.Sprintf("%s%s%s", "http://", lockhost, "/pcc/closelock")
	res := types.GetLockRes{LockMsgList: lockmsglist}

	resp, err := httpc.Do(context.Background(), http.MethodPost, urlPath, res)
	if err != nil {
		fmt.Println(err)
	}
	if resp == nil || resp.Body == nil {
		return false
	}
	lockresult := &types.GetLockResp{Code: make(map[string]bool)}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, lockresult)
	defer resp.Body.Close()
	for _, b := range lockresult.Code {
		if b == false {
			return false
		}
	}
	return true

}
func (l *CoupontellmesoLogic) oplog(tablename, event, describe string, lid int64) error {
	aol := &cachemodel.AccountOperateLog{Phone: l.ctx.Value("phone").(string), TableName: tablename, Event: event, Describe: describe, Timestamp: time.Now(), Lid: lid}
	_, err := l.svcCtx.AccountOperateLog.Insert(l.ctx, aol)
	return err
}
