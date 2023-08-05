package orderpay

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"net/http"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"time"
)

type UtilLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	phone  string
}

func NewUtilLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UtilLogic {
	return &UtilLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		phone:  ctx.Value("phone").(string),
	}
}
func md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	//return hex.EncodeToString(h.Sum(nil))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
func (l *UtilLogic) Oplog(tablename, event, describe string, lid int64) error {
	aol := &cachemodel.AccountOperateLog{Phone: l.phone, TableName: tablename, Event: event, Describe: describe, Timestamp: time.Now(), Lid: lid}
	_, err := l.svcCtx.AccountOperateLog.Insert(l.ctx, aol)
	return err
}
func (l *UtilLogic) Getlocktry(lockmsglist []*types.LockMsg) bool {
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
func (l *UtilLogic) getlocksingletry(lockmsglist []*types.LockMsg) bool {
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
func (l *UtilLogic) Closelock(lockmsglist []*types.LockMsg) bool {
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
func int2bool(a int64) bool {
	if a == 0 {
		return false
	} else {
		return true
	}
}
func bool2int(yes bool) int64 {
	if yes {
		return 1
	} else {
		return 0
	}
}
func getsha512(message string) string {
	bytes2 := sha512.Sum512([]byte(message))   //计算哈希值，返回一个长度为32的数组
	hashCode2 := hex.EncodeToString(bytes2[:]) //将数组转换成切片，转换成16进制，返回字符串
	return hashCode2
}
func GetSha256(message string) string {
	bytes2 := sha256.Sum256([]byte(message))   //计算哈希值，返回一个长度为32的数组
	hashCode2 := hex.EncodeToString(bytes2[:]) //将数组转换成切片，转换成16进制，返回字符串
	return hashCode2
}
func getbool(intbool int64) bool {
	if intbool == 0 {
		return false
	}
	return true
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
func OrderCanBeOvertime(order *cachemodel.Order, PayInfo *cachemodel.PayInfo) bool {
	if order.OrderStatus == 0 {
		//over:=order.CreateOrderTime.Add(time.Minute*15).Before(time.Now())
		over := order.CreateOrderTime.Add(time.Second * 30).Before(time.Now())
		if !PartPay(PayInfo) && over {
			return true
		} else {
			return false
		}
	}
	return false
}
func OrderCanBeDeleted(order *cachemodel.Order, PayInfo *cachemodel.PayInfo) bool {
	if order.OrderStatus == 7 || order.OrderStatus == 3 {
		return true
	}
	if order.OrderStatus == 0 {
		if !PartPay(PayInfo) {
			return true
		} else {
			return false
		}
	}
	return false
}

func PartPay(PayInfo *cachemodel.PayInfo) bool {
	cash := false
	weixin := false
	if PayInfo.CashAccountPayAmount > 0 && PayInfo.FinishAccountpay == 1 {
		cash = true
	}
	if PayInfo.WexinPayAmount > 0 && PayInfo.FinishWeixinpay == 1 {
		weixin = true
	}
	if cash == false && weixin == false {
		return false
	} else {
		return true
	}
}
func couponfinish(order *cachemodel.Order, usercoupon *cachemodel.UserCoupon, use bool) (bool, *cachemodel.UserCoupon) {
	ok, _, ucm := CouponInfoDeal(order, usercoupon, use)
	marshal, _ := json.Marshal(ucm)
	usercoupon.CouponIdMap = string(marshal)
	return ok, usercoupon
}
func CouponInfoDeal(order *cachemodel.Order, usercoupon *cachemodel.UserCoupon, use bool) (bool, *types.CouponStoreInfo, map[int64]map[string]*types.CouponStoreInfo) {
	usercouponmap := make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(usercoupon.CouponIdMap), &usercouponmap)
	if use {
		return deletkv(order, usercouponmap)
	} else {
		return addkv(order, usercouponmap)
	}
}
func deletkv(order *cachemodel.Order, usercouponmap map[int64]map[string]*types.CouponStoreInfo) (bool, *types.CouponStoreInfo, map[int64]map[string]*types.CouponStoreInfo) {
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
func addkv(order *cachemodel.Order, usercouponmap map[int64]map[string]*types.CouponStoreInfo) (bool, *types.CouponStoreInfo, map[int64]map[string]*types.CouponStoreInfo) {
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
func ordercoupondetail(order *cachemodel.Order) (int64, string, *types.CouponStoreInfo) {
	usercouponmap := make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(order.UsedCouponinfo), &usercouponmap)
	for couponid, v := range usercouponmap {
		for uuid, info := range v {
			return couponid, uuid, info
		}
	}
	return 0, "", nil
}
func OrderDb2info(order *cachemodel.Order) *types.OrderInfo {
	FreightAmount := int64(3000)
	CutFreightAmount := int64(3000)
	if FreightAmount-order.FreightAmount > 0 {
		CutFreightAmount = FreightAmount - order.FreightAmount
	} else {
		FreightAmount = order.FreightAmount
		CutFreightAmount = 0
	}
	orderinfo := &types.OrderInfo{}
	orderinfo.Phone = order.Phone
	orderinfo.OrderSn = order.OrderSn
	orderinfo.OutTradeNo = order.OutTradeNo
	orderinfo.PointsOrder = getbool(order.PointsOrder)
	orderinfo.UsedPoints = order.PointsAmount
	orderinfo.CreateTime = order.CreateOrderTime.Format("2006-01-02 15:04:05")
	OrderProducInfo := make([]*types.OrderProductInfo, 0)
	json.Unmarshal([]byte(order.Pidlist), &OrderProducInfo)

	orderinfo.ProductInfo = OrderProducInfo
	orderinfo.UsedCouponInfo = storedcouponinfo2typeinfo(order.UsedCouponinfo)
	orderinfo.ProductCutAmount = float64(order.OriginalAmount-order.PromotionAmount) / 100
	orderinfo.PromotionAmount = float64(order.PromotionAmount) / 100
	orderinfo.OriginalAmount = float64(order.OriginalAmount) / 100
	orderinfo.ActualAmount = float64(order.ActualAmount) / 100
	orderinfo.CouponAmount = float64(order.CouponAmount) / 100
	orderinfo.WeXinPayAmount = float64(order.WexinPayAmount) / 100
	orderinfo.InvoiceAmount = orderinfo.WeXinPayAmount
	orderinfo.CashAccountPayAmount = float64(order.CashAccountPayAmount) / 100
	//orderinfo.FreightAmount = float64(order.FreightAmount) / 100
	orderinfo.FreightAmount = float64(FreightAmount) / 100
	orderinfo.CutFreightAmount = float64(CutFreightAmount) / 100
	orderinfo.RealFreightAmount = float64(order.FreightAmount) / 100
	orderinfo.IfCutFreight = true
	orderinfo.CutPrice = float64(order.OriginalAmount-order.WexinPayAmount+order.FreightAmount) / 100
	orderinfo.CutPriceWithFreight = orderinfo.CutPrice + orderinfo.CutFreightAmount
	orderinfo.Growth = order.Growth
	orderinfo.InvoiceStatus = order.InvoiceStatus
	orderinfo.OrderStatus = order.OrderStatus
	orderinfo.DeliveryCompany = order.DeliveryCompany
	orderinfo.DeliverySn = order.DeliverySn
	orderinfo.MarketPlayerId = order.MarketPlayerId

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
	if order.OrderStatus == 2 {
		orderinfo.RouteList = GetRoutesList(order.DeliverySn)
	}
	if order.OrderStatus == 3 && order.ReceiveTime.Before(time.Now().Add(time.Hour*24*2)) && order.ReceiveTime.Add(time.Hour*24*2).After(time.Now()) {

		orderinfo.RouteList = GetRoutesList(order.DeliverySn)
	}
	return orderinfo
}
func cashfinish(CashAccountPayAmount int64, cashaccount *cachemodel.CashAccount, use bool) (*cachemodel.CashAccount, bool) {
	if use {
		if (cashaccount.Balance - CashAccountPayAmount) < 0 {
			return cashaccount, false
		} else {
			cashaccount.Balance = cashaccount.Balance - CashAccountPayAmount
			return cashaccount, true
		}
	} else {
		cashaccount.Balance = cashaccount.Balance + CashAccountPayAmount
		return cashaccount, true
	}
}
func status2key(status int64) int64 {

	switch status {
	case 1, 3:
		return 1
	case 2, 4:
		return 2
	}
	return 0
}
func Uuidstr2map(str string) (uuidmap map[int64]map[string]*types.CouponStoreInfo) {
	uuidmap = make(map[int64]map[string]*types.CouponStoreInfo)
	json.Unmarshal([]byte(str), &uuidmap)
	return uuidmap
}
func ifcouponuseable(coupon *cachemodel.Coupon, disableTime string, amount int64) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	disabletime, _ := time.Parse("2006-01-02 15:04:05", disableTime)
	if time.Now().After(disabletime) {
		return false
	}
	if coupon.MinPoint != 0 && coupon.Cut != 0 {
		if amount >= coupon.MinPoint {
			return true
		}
	}
	return false
}
func getQuantityBool(cut int64) bool {
	if cut > 0 {
		return true
	} else {
		return false
	}
}
func getIfReserve(status int64) bool {
	switch status {
	case 2:
		return true
	case 4:
		return true
	}
	return false
}
