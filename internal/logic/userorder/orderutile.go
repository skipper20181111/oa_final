package userorder

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/mathx"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"math/rand"
	"net/http"
	"oa_final/cachemodel"
	"oa_final/internal/types"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/internal/svc"
)

type Logic struct {
	logx.Logger
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	Orderdb       *cachemodel.UserOrder
	cashaccount   *cachemodel.CashAccount
	coupon        *cachemodel.Coupon
	usercoupon    *cachemodel.UserCoupon
	userpoints    *cachemodel.UserPoints
	userphone     string
	usecoupon     bool
	usepoint      bool
	usecash       bool
	useropenid    string
	couponid      int64
	couponinfomap map[int64]map[string]*types.CouponStoreInfo
	couponuuid    string
}

func NewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Logic {
	return &Logic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
	}
}
func (l *Logic) WeixinPay() {

}
func OrderDb2Preinfo(order *cachemodel.UserOrder) *types.PreOrderInfo {
	orderinfo := &types.PreOrderInfo{}
	orderinfo.Phone = order.Phone
	orderinfo.PointAmount = order.PointAmount
	orderinfo.CreateTime = order.CreateOrderTime.Format("2006-01-02 15:04:05")
	pidlist := make([]*types.ProductTiny, 0)
	json.Unmarshal([]byte(order.Pidlist), &pidlist)
	orderinfo.PidList = pidlist
	orderinfo.OriginalAmount = float64(order.OriginalAmount) / 100
	orderinfo.ActualAmount = float64(order.ActualAmount) / 100
	orderinfo.CouponAmount = float64(order.CouponAmount) / 100
	orderinfo.WeXinPayAmount = float64(order.WexinPayAmount) / 100
	orderinfo.CashAccountPayAmount = float64(order.CashAccountPayAmount) / 100
	orderinfo.FreightAmount = float64(order.FreightAmount) / 100
	return orderinfo
}
func Uuidstr2map(str string) (uuidmap map[int64]map[string]*types.CouponStoreInfo) {
	uuidmap = make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(str), &uuidmap)
	return uuidmap
}
func storedcouponinfo2typeinfo(infostr string) *types.CouponStoreInfo {
	couponinfo := make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(infostr), &couponinfo)
	for _, V := range couponinfo {
		for _, info := range V {
			return info
		}
	}
	return nil
}
func OrderDb2info(order *cachemodel.UserOrder, info *cachemodel.TransactionInfo) *types.OrderInfo {

	orderinfo := &types.OrderInfo{}
	if info != nil {
		orderinfo.CashAccountPayAmount = float64(info.CashAccountPayAmount) / 100
		orderinfo.WeXinPayAmount = float64(info.WexinPayAmount) / 100
	}

	orderinfo.Phone = order.Phone
	orderinfo.PointAmount = float64(order.PointAmount) / 100
	orderinfo.OrderSn = order.OrderSn
	orderinfo.OutTradeNo = order.OutTradeNo
	orderinfo.TransactionId = order.TransactionId
	orderinfo.CreateTime = order.CreateOrderTime.Format("2006-01-02 15:04:05")
	pidlist := make([]*types.ProductTiny, 0)
	json.Unmarshal([]byte(order.Pidlist), &pidlist)
	orderinfo.PidList = pidlist
	orderinfo.UsedCouponInfo = storedcouponinfo2typeinfo(order.UsedCouponinfo)
	orderinfo.ProductCutAmount = float64(order.OriginalOriginalAmount-order.OriginalAmount) / 100
	orderinfo.OriginalAmount = float64(order.OriginalOriginalAmount) / 100
	orderinfo.ActualAmount = float64(order.ActualAmount) / 100
	orderinfo.CouponAmount = float64(order.CouponAmount) / 100
	orderinfo.WeXinPayAmount = float64(order.WexinPayAmount) / 100
	orderinfo.CashAccountPayAmount = float64(order.CashAccountPayAmount) / 100
	orderinfo.FreightAmount = float64(order.FreightAmount) / 100
	orderinfo.CutFreightAmount = 10
	orderinfo.RealFreightAmount = orderinfo.FreightAmount - orderinfo.CutFreightAmount
	orderinfo.IfCutFreight = true
	orderinfo.CutPrice = float64(order.OriginalOriginalAmount-order.WexinPayAmount) / 100
	orderinfo.Growth = order.Growth

	orderinfo.OrderStatus = order.OrderStatus
	orderinfo.DeliveryCompany = order.DeliveryCompany
	orderinfo.DeliverySn = order.DeliverySn

	address := types.AddressInfo{}
	json.Unmarshal([]byte(order.Address), &address)
	orderinfo.Address = &address
	orderinfo.OrderNote = order.OrderNote
	orderinfo.ConfirmStatus = order.ConfirmStatus
	orderinfo.DeleteStatus = order.DeleteStatus
	orderinfo.PaymentTime = order.PaymentTime.Format("2006-01-02 15:04:05")
	orderinfo.ModifyTime = order.ModifyTime.Format("2006-01-02 15:04:05")
	orderinfo.DeliveryTime = order.DeliveryTime.Format("2006-01-02 15:04:05")
	orderinfo.ReceiveTime = order.ReceiveTime.Format("2006-01-02 15:04:05")
	return orderinfo
}
func (l *Logic) Order2db(req *types.NewOrderRes, productsMap map[int64]*cachemodel.Product, opts ...func(logic *Logic)) *cachemodel.UserOrder {
	l.couponid = req.UsedCouponId
	l.couponinfomap = Uuidstr2map(req.UsedCouponUUID)
	l.couponuuid = req.UsedCouponUUID
	for _, option := range opts {
		option(l)
	}
	inittime, _ := time.Parse("2006-01-02 15:04:05", "2099-01-01 00:00:00")
	order := &cachemodel.UserOrder{}
	order.Phone = l.userphone
	order.FinishWeixinpay = 0
	order.PointsOrder = 0
	order.FinishAccountpay = 0
	order.CreateOrderTime = time.Now()
	order.OutTradeNo = randStr(32)
	marshal, err := json.Marshal(req.ProductTinyList)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	order.Pidlist = string(marshal)
	for _, tiny := range req.ProductTinyList {
		order.OriginalAmount = order.OriginalAmount + productsMap[tiny.PId].PromotionPrice*int64(tiny.Amount)
	}
	for _, tiny := range req.ProductTinyList {
		order.OriginalOriginalAmount = order.OriginalOriginalAmount + productsMap[tiny.PId].OriginalPrice*int64(tiny.Amount)
	}
	order.ActualAmount = order.OriginalAmount
	l.Orderdb = order
	l.calculatemoney(req.UseCouponFirst, req.UseCashFirst, opts...)
	if req.UsePointFirst {
		if l.userpoints != nil && l.userpoints.AvailablePoints > 0 {
			l.usepoint = true
			order.PointAmount = int64(mathx.MinInt(int(order.OriginalAmount), int(l.userpoints.AvailablePoints)))
			order.ActualAmount = order.ActualAmount - order.PointAmount
		}
	}
	order.FreightAmount = 4000
	order.OrderStatus = 0
	order.DeliveryCompany = "顺丰"
	order.DeliverySn = randStr(20)
	addr, err := json.Marshal(req.Address)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	order.Address = string(addr)
	//order.BillType = 0
	//order.BillInfo =""

	order.OrderNote = req.OrderNote
	order.DeleteStatus = 0
	order.Growth = order.ActualAmount
	order.ConfirmStatus = 0
	order.ModifyTime = order.CreateOrderTime
	order.PaymentTime = inittime
	order.DeliveryTime = inittime
	order.ReceiveTime = inittime
	order.CloseTime = inittime
	order.OrderSn = getsha512(order.Phone + order.CreateOrderTime.String() + order.Pidlist + order.Address)
	order.LogId = time.Now().UnixNano()
	l.Orderdb = order
	return order
}
func getsha512(message string) string {
	bytes2 := sha512.Sum512([]byte(message))   //计算哈希值，返回一个长度为32的数组
	hashCode2 := hex.EncodeToString(bytes2[:]) //将数组转换成切片，转换成16进制，返回字符串
	return hashCode2
}

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
func (l *Logic) Oplog(tablename, event, describe string, lid int64) error {
	aol := &cachemodel.AccountOperateLog{Phone: l.ctx.Value("phone").(string), TableName: tablename, Event: event, Describe: describe, Timestamp: time.Now(), Lid: lid}
	_, err := l.svcCtx.AccountOperateLog.Insert(l.ctx, aol)
	return err
}
func (l *Logic) coupondb2storeinfo() string {
	couponstoreinfomapori := make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(l.usercoupon.CouponIdMap), &couponstoreinfomapori)
	storemap := make(map[int64]map[string]*types.CouponStoreInfo)
	storemap[l.couponid] = make(map[string]*types.CouponStoreInfo)
	storemap[l.couponid][l.couponuuid] = couponstoreinfomapori[l.couponid][l.couponuuid]
	marshal, _ := json.Marshal(storemap)
	return string(marshal)
}
func (l *Logic) couponeeffective() bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	if l.coupon != nil && l.usercoupon != nil {
		usercouponmap := make(map[int64]map[string]*types.CouponStoreInfo)
		json.Unmarshal([]byte(l.usercoupon.CouponIdMap), &usercouponmap)
		_, ok := usercouponmap[l.couponid] //连续两次判断我是否有这个优惠券
		if ok {
			_, ok := usercouponmap[l.couponid][l.couponuuid]
			if ok {
				disabledtime, _ := time.Parse("2006-01-02 15:04:05", usercouponmap[l.couponid][l.couponuuid].DisabledTime)
				if disabledtime.After(time.Now()) {
					return true
				}
			}
		}
	}
	return false
}
func (l *Logic) calculatemoney(UseCoupon, usecash bool, options ...func(logic *Logic)) *cachemodel.UserOrder {
	l.usecoupon = false
	l.Orderdb.UsedCouponinfo = ""
	l.Orderdb.CouponAmount = 0
	if UseCoupon {
		//计算打折后的钱
		if l.couponeeffective() {
			if l.coupon.Discount != 0 {
				l.usecoupon = true
				discountammount := int64(float64(l.Orderdb.ActualAmount) * float64(l.coupon.Discount) / 100)
				l.Orderdb.UsedCouponinfo = l.coupondb2storeinfo()
				l.Orderdb.CouponAmount = l.Orderdb.ActualAmount - discountammount
				l.Orderdb.ActualAmount = discountammount

			} else if l.coupon.MinPoint != 0 && l.coupon.Cut != 0 {
				if l.Orderdb.ActualAmount >= l.coupon.MinPoint {
					l.usecoupon = true
					l.Orderdb.UsedCouponinfo = l.coupondb2storeinfo()
					l.Orderdb.CouponAmount = l.coupon.Cut
					l.Orderdb.ActualAmount = l.Orderdb.ActualAmount - l.Orderdb.CouponAmount
				}
			}
		}

	} else {
		l.Orderdb.CouponAmount = 0
	}

	// usecash 暂时不用了
	//if usecash {
	//	cash := l.cashaccount
	//	if cash != nil {
	//		if cash.Balance*100 > 0 {
	//			l.usecash = true
	//			if (l.Orderdb.ActualAmount - cash.Balance) >= 0 {
	//				l.Orderdb.WexinPayAmount = l.Orderdb.ActualAmount - cash.Balance*100
	//				l.Orderdb.CashAccountPayAmount = cash.Balance
	//			} else {
	//				l.Orderdb.WexinPayAmount = 0
	//				l.Orderdb.CashAccountPayAmount = l.Orderdb.ActualAmount
	//			}
	//		} else {
	//			l.Orderdb.WexinPayAmount = l.Orderdb.ActualAmount
	//			l.Orderdb.CashAccountPayAmount = 0
	//		}
	//
	//	} else {
	//		l.Orderdb.WexinPayAmount = l.Orderdb.ActualAmount
	//	}
	//} else {
	//	l.Orderdb.WexinPayAmount = l.Orderdb.ActualAmount
	//}

	return l.Orderdb
}
func (l *Logic) Getlocktry(lockmsglist []*types.LockMsg) bool {
	trytime := 3
	for i := 0; i < trytime; i++ {
		ok := l.getlocksingletry(lockmsglist)
		if ok {
			return ok
		} else {
			time.Sleep(time.Second)
		}
	}
	return false
}
func (l *Logic) getlocksingletry(lockmsglist []*types.LockMsg) bool {
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
func (l *Logic) Closelock(lockmsglist []*types.LockMsg) bool {
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

type Option func(*Logic)

func UseCache(usecash bool) Option {
	return func(l *Logic) {
		if usecash {
			l.coupon, _ = l.svcCtx.Coupon.FindOneByCouponId(l.ctx, l.couponid)
			l.usercoupon, _ = l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, l.userphone)
			l.cashaccount, _ = l.svcCtx.CashAccount.FindOneByPhone(l.ctx, l.userphone)
			l.userpoints, _ = l.svcCtx.UserPoints.FindOneByPhone(l.ctx, l.userphone)
		} else {
			l.coupon, _ = l.svcCtx.Coupon.FindOneByCouponIdNoCache(l.ctx, l.couponid)
			l.usercoupon, _ = l.svcCtx.UserCoupon.FindOneByPhoneNoCache(l.ctx, l.userphone)
			//l.cashaccount, _ = l.svcCtx.CashAccount.FindOneByPhoneNoCach(l.ctx, l.userphone)
			l.userpoints, _ = l.svcCtx.UserPoints.FindOneByPhoneNoCache(l.ctx, l.userphone)
		}
	}
}

func (l *Logic) Updatecashaccount(order *cachemodel.UserOrder, use bool) (bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	accphone := order.Phone
	cashaccount, _ := l.svcCtx.CashAccount.FindOneByPhoneNoCach(l.ctx, accphone)
	account, ok := cashfinish(order, cashaccount, use)
	if ok {
		l.Oplog("cash_account", order.OrderSn, "开始更新", order.LogId)
		l.svcCtx.CashAccount.Update(l.ctx, account)
		l.Oplog("cash_account", order.OrderSn, "结束更新", order.LogId)
		if use {
			l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), Behavior: "消费", Phone: accphone, Balance: cashaccount.Balance, ChangeAmount: order.CashAccountPayAmount})
			l.svcCtx.TransactionInfo.UpdateCashPay(l.ctx, order.OrderSn)
		} else {
			l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), Behavior: "退款", Phone: accphone, Balance: cashaccount.Balance, ChangeAmount: order.CashAccountPayAmount})
			l.svcCtx.TransactionInfo.UpdateCashReject(l.ctx, order.OrderSn)
		}
		return ok, "yes"
	} else {
		return ok, "no"
	}

}
func cashfinish(order *cachemodel.UserOrder, cashaccount *cachemodel.CashAccount, use bool) (*cachemodel.CashAccount, bool) {
	if use {
		if (cashaccount.Balance - order.CashAccountPayAmount) < 0 {
			return cashaccount, false
		} else {
			cashaccount.Balance = cashaccount.Balance - order.CashAccountPayAmount
			return cashaccount, true
		}
	} else {
		cashaccount.Balance = cashaccount.Balance + order.CashAccountPayAmount
		return cashaccount, true
	}
}
func (l *Logic) UpdateCoupon(order *cachemodel.UserOrder, use bool) (bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	accphone := l.ctx.Value("phone").(string)
	usercoupon, _ := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, accphone)
	//l.oplog("usercounpon", order.OrderSn, "开始更新", order.LogId)
	ok, coupon := couponfinish(order, usercoupon, use)
	if ok {
		l.svcCtx.UserCoupon.Update(l.ctx, coupon)
		return ok, "yes"
	} else {
		return ok, "no"
	}
	//l.oplog("usercounpon", order.OrderSn, "结束更新", order.LogId)

}
func couponfinish(order *cachemodel.UserOrder, usercoupon *cachemodel.UserCoupon, use bool) (bool, *cachemodel.UserCoupon) {
	ok, _, ucm := CouponInfoDeal(order, usercoupon, use)
	marshal, _ := json.Marshal(ucm)
	usercoupon.CouponIdMap = string(marshal)
	return ok, usercoupon
}
func CouponInfoDeal(order *cachemodel.UserOrder, usercoupon *cachemodel.UserCoupon, use bool) (bool, *types.CouponStoreInfo, map[int64]map[string]*types.CouponStoreInfo) {
	usercouponmap := make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(usercoupon.CouponIdMap), &usercouponmap)
	if use {
		return deletkv(order, usercouponmap)
	} else {
		return addkv(order, usercouponmap)
	}
}
func ordercoupondetail(order *cachemodel.UserOrder) (int64, string, *types.CouponStoreInfo) {
	usercouponmap := make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(order.UsedCouponinfo), &usercouponmap)
	for couponid, v := range usercouponmap {
		for uuid, info := range v {
			return couponid, uuid, info
		}
	}
	return 0, "", nil
}
func deletkv(order *cachemodel.UserOrder, usercouponmap map[int64]map[string]*types.CouponStoreInfo) (bool, *types.CouponStoreInfo, map[int64]map[string]*types.CouponStoreInfo) {
	cid, uuid, info := ordercoupondetail(order)
	m, ok := usercouponmap[cid]
	if ok {
		_, ok := m[uuid]
		if ok {
			delete(usercouponmap[cid], uuid)
			return true, info, usercouponmap
		}
	}
	return false, nil, usercouponmap
}
func addkv(order *cachemodel.UserOrder, usercouponmap map[int64]map[string]*types.CouponStoreInfo) (bool, *types.CouponStoreInfo, map[int64]map[string]*types.CouponStoreInfo) {
	cid, uuid, info := ordercoupondetail(order)
	_, ok := usercouponmap[cid]
	if ok {
		usercouponmap[cid][uuid] = info
	} else {
		usercouponmap[cid] = make(map[string]*types.CouponStoreInfo)
		usercouponmap[cid][uuid] = info
	}
	return true, info, usercouponmap
}
func HaveKey(orderpart, couponpart string) bool {
	ordermap := make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(orderpart), &ordermap)
	couponmap := make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(couponpart), &couponmap)
	if len(ordermap) <= 0 {
		return false
	}
	var key int64
	var uuid string
	for k, m := range ordermap {
		key = k
		if len(m) <= 0 {
			return false
		}
	}
	m, ok := couponmap[key]
	if ok {
		_, ok := m[uuid]
		if ok {
			return true
		}
	}
	return false
}

func order2req(order *cachemodel.UserOrder) *types.NewOrderRes {
	req := &types.NewOrderRes{}
	pidlist := make([]*types.ProductTiny, 0)
	json.Unmarshal([]byte(order.Pidlist), &pidlist)
	req.ProductTinyList = pidlist
	req.Address = &types.AddressInfo{}
	if order.CashAccountPayAmount > 0 {
		req.UseCashFirst = true
	}
	if order.PointAmount > 0 {
		req.UsePointFirst = true
	}
	if order.UsedCouponinfo != "" {
		uuidmap := Uuidstr2map(order.UsedCouponinfo)
		for idk, couponmap := range uuidmap {
			for uuidk, _ := range couponmap {
				req.UseCouponFirst = true
				req.UsedCouponId = idk
				req.UsedCouponUUID = uuidk
			}
		}
	}
	return req
}
