package userorder

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelorderLogic {
	return &CancelorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelorderLogic) Cancelorder(req *types.CancelOrderRes) (resp *types.CancelOrderResp, err error) {
	if l.ctx.Value("openid") != req.OpenId || l.ctx.Value("phone") != req.Phone {
		return &types.CancelOrderResp{
			Code: "4004",
			Msg:  "请勿使用其他用户的token",
		}, nil
	}
	lid := time.Now().UnixNano() + int64(rand.Intn(1024))
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	if sn2order == nil {
		return &types.CancelOrderResp{Code: "10000", Msg: "未查询到订单"}, nil
	}
	if sn2order.OrderStatus != 1 {
		return &types.CancelOrderResp{Code: "10000", Msg: "已发货无法退款"}, nil
	}
	l.oplog("微信退款", req.Phone, "开始更新", lid)
	service := refunddomestic.RefundsApiService{Client: l.svcCtx.Client}
	create, result, err := service.Create(l.ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(sn2order.OutTradeNo),
		OutRefundNo: core.String(sn2order.OutTradeNo),
		Amount: &refunddomestic.AmountReq{Currency: core.String("CNY"),
			Refund: core.Int64(sn2order.WexinPayAmount),
			Total:  core.Int64(sn2order.WexinPayAmount)},
	})
	defer result.Response.Body.Close()
	if err != nil {
		log.Printf("call Create err:%s", err)
		return &types.CancelOrderResp{Code: "4004", Msg: err.Error()}, nil
	} else {
		log.Printf("status=%d resp=%s", result.Response.StatusCode, resp, create.String())
	}
	l.oplog("微信退款", req.Phone, "结束更新", lid)
	//从这里开始更新现金账户于优惠券账户
	// 此时还有特别重要的事情，1，要更改现金账户余额，2，要更改优惠券账户，毕竟优惠券账户已经用完了。
	if sn2order.CashAccountPayAmount > 0 || sn2order.UsedCouponid != 0 {
		lockmsglist := make([]*types.LockMsg, 0)
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.ctx.Value("phone").(string), Field: "user_coupon"})
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.ctx.Value("phone").(string), Field: "cash_account"})
		if l.getlock(lockmsglist) {
			if sn2order.CashAccountPayAmount > 0 && !l.updatecashaccount(lid, sn2order) {
				l.oplog("更新现金账户失败", req.Phone+sn2order.OutTradeNo, "开始更新", lid)
			}
			if sn2order.UsedCouponid != 0 && !l.updatecoupon(lid, sn2order) {
				l.oplog("更新优惠券失败", req.Phone+sn2order.OutTradeNo, "开始更新", lid)
			}
		}
	}

	//结束更新现金账户与优惠券账户

	sn2order.OrderStatus = 6
	sn2order.ModifyTime = time.Now()
	err = l.svcCtx.UserOrder.Update(l.ctx, sn2order)
	sn, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, req.OrderSn)
	return &types.CancelOrderResp{Code: "10000", Msg: "发起退款成功", Data: &types.CancelOrderRp{OrderInfo: db2orderinfo(sn)}}, nil
}

func (l *CancelorderLogic) oplog(tablename, event, describe string, lid int64) error {
	aol := &cachemodel.AccountOperateLog{Phone: l.ctx.Value("phone").(string), TableName: tablename, Event: event, Describe: describe, Timestamp: time.Now(), Lid: lid}
	_, err := l.svcCtx.AccountOperateLog.Insert(l.ctx, aol)
	return err
}
func (l *CancelorderLogic) getlock(lockmsglist []*types.LockMsg) bool {
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
func (l *CancelorderLogic) closelock(lockmsglist []*types.LockMsg) bool {
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
func (l *CancelorderLogic) updatecashaccount(lid int64, order *cachemodel.UserOrder) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()

	accphone := l.ctx.Value("phone").(string)
	phone, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, accphone)
	l.oplog("cash_account", accphone, "开始更新", lid)
	phone.Balance = phone.Balance + float64(order.CashAccountPayAmount)/100
	l.svcCtx.CashAccount.Update(l.ctx, phone)
	l.oplog("cash_account", accphone, "结束更新", lid)
	return true
}
func (l *CancelorderLogic) updatecoupon(lid int64, order *cachemodel.UserOrder) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	accphone := l.ctx.Value("phone").(string)
	phone, _ := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, accphone)
	l.oplog("usercounpon", accphone, "开始更新", lid)
	usercouponmap := make(map[int64]int)
	json.Unmarshal([]byte(phone.CouponIdList), &usercouponmap)
	usercouponmap[order.UsedCouponid] = usercouponmap[order.UsedCouponid] + 1
	marshal, _ := json.Marshal(usercouponmap)
	phone.CouponIdList = string(marshal)
	l.svcCtx.UserCoupon.Update(l.ctx, phone)
	l.oplog("usercounpon", accphone, "结束更新", lid)
	return true
}
