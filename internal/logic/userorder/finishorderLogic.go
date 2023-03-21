package userorder

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"net/http"
	"oa_final/cachemodel"
	"time"

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
	l.usecoupon = false
	l.usecash = false
	l.usepoint = false
	userphone := l.ctx.Value("phone").(string)

	//从这里开始更新现金账户于优惠券账户
	// 此时还有特别重要的事情，1，要更改现金账户余额，2，要更改优惠券账户，毕竟优惠券账户已经用完了。
	order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	l.userorder = order
	if order.CashAccountPayAmount > 0 {
		l.usecash = true
	}
	if order.UsedCouponid != -1 {
		l.usecoupon = true
	}
	if order.PointAmount > 0 {
		cache, _ := l.svcCtx.UserPoints.FindOneByPhoneNoCache(l.ctx, userphone)
		if cache != nil {
			cache.AvailablePoints = cache.AvailablePoints - order.PointAmount
			l.svcCtx.UserPoints.Update(l.ctx, cache)
		}
	}
	lid := order.LogId
	if l.usecash || l.usecoupon {
		lockmsglist := make([]*types.LockMsg, 0)
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: userphone, Field: "user_coupon"})
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: userphone, Field: "cash_account"})
		if l.getlock(lockmsglist) {
			if l.usecash {
				if !l.updatecashaccount(lid) {
					l.oplog("支付模块更新现金账户失败", order.OrderSn, "开始更新", lid)
				}
			}

			if l.usecoupon {
				if !l.updatecoupon(lid) {
					l.oplog("支付模块更新优惠券失败", order.OrderSn, "开始更新", lid)
				}
			}

		}
		l.closelock(lockmsglist)
	}

	return &types.FinishOrderResp{Code: "10000", Msg: "finished", Data: db2orderinfo(order)}, nil
}
func (l *FinishorderLogic) oplog(tablename, event, describe string, lid int64) error {
	aol := &cachemodel.AccountOperateLog{Phone: l.ctx.Value("phone").(string), TableName: tablename, Event: event, Describe: describe, Timestamp: time.Now(), Lid: lid}
	_, err := l.svcCtx.AccountOperateLog.Insert(l.ctx, aol)
	return err
}
func (l *FinishorderLogic) getlock(lockmsglist []*types.LockMsg) bool {
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
func (l *FinishorderLogic) closelock(lockmsglist []*types.LockMsg) bool {
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
func (l *FinishorderLogic) updatecashaccount(lid int64) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()

	accphone := l.ctx.Value("phone").(string)
	phone, _ := l.svcCtx.CashAccount.FindOneByPhoneNoCach(l.ctx, accphone)
	l.oplog("cash_account", l.userorder.OrderSn, "开始更新", lid)
	phone.Balance = phone.Balance - float64(l.userorder.CashAccountPayAmount)/100
	l.svcCtx.CashAccount.Update(l.ctx, phone)
	l.oplog("cash_account", l.userorder.OrderSn, "结束更新", lid)
	l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), Behavior: "消费", Phone: accphone, Balance: phone.Balance, ChangeAmount: l.cashaccount.Balance})
	return true
}
func (l *FinishorderLogic) updatecoupon(lid int64) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	accphone := l.ctx.Value("phone").(string)
	phone, _ := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, accphone)
	l.oplog("usercounpon", l.userorder.OrderSn, "开始更新", lid)
	usercouponmap := make(map[int64]int)
	json.Unmarshal([]byte(phone.CouponIdList), &usercouponmap)
	usercouponmap[l.userorder.UsedCouponid] = usercouponmap[l.userorder.UsedCouponid] - 1
	marshal, _ := json.Marshal(usercouponmap)
	phone.CouponIdList = string(marshal)
	l.svcCtx.UserCoupon.Update(l.ctx, phone)
	l.oplog("usercounpon", l.userorder.OrderSn, "结束更新", lid)
	return true
}
