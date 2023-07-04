package orderpay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"time"
)

type OrderUtilLogic struct {
	logx.Logger
	ctx                context.Context
	svcCtx             *svc.ServiceContext
	userphone          string
	ProductsMap        map[int64]*cachemodel.Product
	req                *types.NewOrderRes
	ProductTinyList    []*types.ProductTiny
	coupon             *cachemodel.Coupon
	usercoupon         *cachemodel.UserCoupon
	UsedCouponStorInfo map[int64]map[string]*types.CouponStoreInfo
	orderlist          []*cachemodel.Order
	PayInit            *types.PayInit
	ul                 *UtilLogic
	MarketPlayerMap    map[int64]map[int64][]*types.ProductTiny
}

func NewOrderUtilLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderUtilLogic {
	PMcache, ok := svcCtx.LocalCache.Get(svc.ProductsMap)
	ProductsMap := make(map[int64]*cachemodel.Product)
	if ok {
		ProductsMap = PMcache.(map[int64]*cachemodel.Product)
	}
	return &OrderUtilLogic{
		Logger:      logx.WithContext(ctx),
		ctx:         ctx,
		svcCtx:      svcCtx,
		userphone:   ctx.Value("phone").(string),
		ProductsMap: ProductsMap,
		ul:          NewUtilLogic(ctx, svcCtx),
	}
}
func order2req(order *cachemodel.Order) *types.NewOrderRes {
	req := &types.NewOrderRes{}
	pidlist := make([]*types.ProductTiny, 0)
	json.Unmarshal([]byte(order.Pidlist), &pidlist)
	req.ProductTinyList = pidlist

	req.Address = &types.AddressInfo{}
	json.Unmarshal([]byte(order.Address), req.Address)

	req.OrderNote = order.OrderNote

	if order.CashAccountPayAmount > 0 {
		req.UseCashFirst = true
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
func (l OrderUtilLogic) GetAmount(ProductTinyList []*types.ProductTiny) (OriginalAmount int64, PromotionAmount int64) {
	OriginalAmount = int64(0)
	PromotionAmount = int64(0)
	l.ProductTinyList = ProductTinyList
	if !l.ProductTinyListChina() {
		return 0, 0
	}
	for _, m1 := range l.MarketPlayerMap {
		for _, producttinylist := range m1 {
			originalAmount, promotionAmount := l.OriProPrice(producttinylist)
			OriginalAmount = OriginalAmount + originalAmount
			PromotionAmount = PromotionAmount + promotionAmount
		}
	}
	return OriginalAmount, PromotionAmount
}

func (l OrderUtilLogic) req2op(req *types.NewOrderRes) ([]*cachemodel.Order, *types.PayInit, bool) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	l.req = req
	l.ProductTinyList = req.ProductTinyList
	l.PayInit = &types.PayInit{}
	l.PayInit.Phone = l.userphone
	l.PayInit.OutTradeSn = randStr(32)
	l.PayInit.NeedCashAccount = l.req.UseCashFirst
	if l.OrderChina() {
		l.EndPayInit()
		return l.orderlist, l.PayInit, true
	}
	return nil, nil, false
}
func (l OrderUtilLogic) ConsumeInfo(ProductTinyList []*types.ProductTiny) {

}
func (l OrderUtilLogic) EndPayInit() {
	BiggestOrderIndex := 0
	BiggestAmount := int64(0)
	for i, order := range l.orderlist {
		if order.PromotionAmount > BiggestAmount {
			BiggestAmount = order.PromotionAmount
			BiggestOrderIndex = i
		}
	}
	l.orderlist[BiggestOrderIndex].ActualAmount = l.CountCouponPrice(BiggestAmount)
	UsedCouponStorInfoStr, _ := json.Marshal(l.UsedCouponStorInfo)
	l.orderlist[BiggestOrderIndex].UsedCouponinfo = string(UsedCouponStorInfoStr)
	l.orderlist[BiggestOrderIndex].CouponAmount = l.coupon.Cut
	for _, order := range l.orderlist {
		l.PayInit.TotleAmmount = l.PayInit.TotleAmmount + order.ActualAmount
	}

}
func (l OrderUtilLogic) ProductTinyListChina() bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	l.MarketPlayerMap = make(map[int64]map[int64][]*types.ProductTiny)
	for _, tiny := range l.ProductTinyList {
		if _, ok := l.MarketPlayerMap[l.ProductsMap[tiny.PId].MarketPlayerId]; ok {
			l.MarketPlayerMap[l.ProductsMap[tiny.PId].MarketPlayerId] = l.PidListMapChina(l.MarketPlayerMap[l.ProductsMap[tiny.PId].MarketPlayerId], tiny)
		} else {
			l.MarketPlayerMap[l.ProductsMap[tiny.PId].MarketPlayerId] = make(map[int64][]*types.ProductTiny)
			l.MarketPlayerMap[l.ProductsMap[tiny.PId].MarketPlayerId] = l.PidListMapChina(l.MarketPlayerMap[l.ProductsMap[tiny.PId].MarketPlayerId], tiny)
		}
	}
	return true
}
func (l OrderUtilLogic) OrderChina() bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	if !l.ProductTinyListChina() {
		return false
	}
	OrderList := l.PidListMap2OrderMap(l.MarketPlayerMap)
	if len(OrderList) >= 1 {
		l.orderlist = OrderList
		return true
	}
	return false
}
func (l OrderUtilLogic) PidListMap2OrderMap(MarketPlayerMap map[int64]map[int64][]*types.ProductTiny) []*cachemodel.Order {
	OrdersList := make([]*cachemodel.Order, 0)
	for _, PidListMap := range MarketPlayerMap {
		for _, PidList := range PidListMap {
			OrdersList = append(OrdersList, l.PidList2Order(PidList))
		}
	}
	return OrdersList
}
func (l OrderUtilLogic) GetPromotionPrice(Tiny *types.ProductTiny) int64 {
	switch l.ProductsMap[Tiny.PId].Status {
	case 3, 4:
		if l.ProductsMap[Tiny.PId].PromotionPrice >= l.ProductsMap[Tiny.PId].MinPrice {
			return l.ProductsMap[Tiny.PId].PromotionPrice - l.ProductsMap[Tiny.PId].Cut
		}
	}
	return l.ProductsMap[Tiny.PId].PromotionPrice
}
func (l OrderUtilLogic) OriProPrice(ProductTinyList []*types.ProductTiny) (OriginalAmount int64, PromotionAmount int64) {
	OriginalAmount = int64(0)
	PromotionAmount = int64(0)
	for _, tiny := range ProductTinyList {
		OriginalAmount = OriginalAmount + l.ProductsMap[tiny.PId].OriginalPrice
		PromotionAmount = PromotionAmount + l.GetPromotionPrice(tiny)
	}
	return OriginalAmount, PromotionAmount
}
func (l OrderUtilLogic) PidList2Order(ProductTinyList []*types.ProductTiny) *cachemodel.Order {
	order := &cachemodel.Order{}
	order.OrderType = status2key(l.ProductsMap[ProductTinyList[0].PId].Status)
	inittime, _ := time.Parse("2006-01-02 15:04:05", "2099-01-01 00:00:00")
	order.Phone = l.userphone
	order.CreateOrderTime = time.Now()
	marshal, _ := json.Marshal(ProductTinyList)
	order.Pidlist = string(marshal)
	order.OriginalAmount, order.PromotionAmount = l.OriProPrice(ProductTinyList)
	order.ActualAmount = order.PromotionAmount
	addr, err := json.Marshal(l.req.Address)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	order.Address = string(addr)
	order.OrderNote = l.req.OrderNote
	order.ModifyTime = order.CreateOrderTime
	order.PaymentTime = inittime
	order.DeliveryTime = inittime
	order.ReceiveTime = inittime
	order.CloseTime = inittime
	order.OrderSn = getsha512(order.Phone + order.CreateOrderTime.String() + order.Pidlist + order.Address)
	order.LogId = time.Now().UnixNano()
	return order
}
func (l OrderUtilLogic) CountCouponPrice(PromotionPrice int64) (ActualAmount int64) {
	if l.req.UseCouponFirst {
		get, ok := l.svcCtx.LocalCache.Get(svc.CouponMapKey)
		l.usercoupon, _ = l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, l.userphone)
		if ok && l.usercoupon != nil {
			l.coupon, ok = get.(map[int64]*cachemodel.Coupon)[l.req.UsedCouponId]
			if ok && l.CouponEffective() {
				if PromotionPrice >= l.coupon.MinPoint {
					return PromotionPrice - l.coupon.Cut
				}
			}
		}
	}
	return PromotionPrice
}
func (l *OrderUtilLogic) CouponEffective() bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	l.UsedCouponStorInfo = make(map[int64]map[string]*types.CouponStoreInfo)
	if l.coupon != nil && l.usercoupon != nil && (l.coupon.MinPoint != 0 && l.coupon.Cut != 0) {
		usercouponmap := make(map[int64]map[string]*types.CouponStoreInfo)
		json.Unmarshal([]byte(l.usercoupon.CouponIdMap), &usercouponmap)
		_, ok := usercouponmap[l.req.UsedCouponId] //连续两次判断我是否有这个优惠券
		if ok {
			_, ok := usercouponmap[l.req.UsedCouponId][l.req.UsedCouponUUID]
			if ok {
				disabledtime, _ := time.Parse("2006-01-02 15:04:05", usercouponmap[l.req.UsedCouponId][l.req.UsedCouponUUID].DisabledTime)
				if disabledtime.After(time.Now()) {
					l.UsedCouponStorInfo[l.coupon.CouponId] = make(map[string]*types.CouponStoreInfo)
					l.UsedCouponStorInfo[l.req.UsedCouponId][l.req.UsedCouponUUID] = usercouponmap[l.req.UsedCouponId][l.req.UsedCouponUUID]
					return true
				}
			}
		}
	}
	return false
}

func (l OrderUtilLogic) PidListMapChina(ProductTinyListMap map[int64][]*types.ProductTiny, Tiny *types.ProductTiny) map[int64][]*types.ProductTiny {
	key := status2key(l.ProductsMap[Tiny.PId].Status)
	if _, ok := ProductTinyListMap[key]; ok {
		ProductTinyListMap[key] = append(ProductTinyListMap[key], Tiny)
	} else {
		ProductTinyListMap[key] = make([]*types.ProductTiny, 0)
		ProductTinyListMap[key] = append(ProductTinyListMap[key], Tiny)
	}
	return ProductTinyListMap
}
func (l *OrderUtilLogic) Updatecashaccount(order *cachemodel.Order, use bool) (bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	accphone := order.Phone
	cashaccount, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, accphone)
	account, ok := cashfinish(order, cashaccount, use)
	if ok {
		l.ul.Oplog("cash_account", order.OrderSn, "开始更新", order.LogId)
		l.svcCtx.CashAccount.Update(l.ctx, account)
		if use {
			l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), OrderType: "消费", OrderSn: order.OrderSn, OrderDescribe: "购买店铺商品消费", Behavior: "消费", Phone: accphone, Balance: cashaccount.Balance, ChangeAmount: order.CashAccountPayAmount})
			l.svcCtx.TransactionInfo.UpdateCashPay(l.ctx, order.OrderSn)
		} else {
			l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), OrderType: "退款", OrderSn: order.OrderSn, OrderDescribe: "店铺商品退款", Behavior: "退款", Phone: accphone, Balance: cashaccount.Balance, ChangeAmount: order.CashAccountPayAmount})
			l.svcCtx.TransactionInfo.UpdateCashReject(l.ctx, order.OrderSn)
		}
		l.ul.Oplog("cash_account", order.OrderSn, "结束更新", order.LogId)
		return ok, "yes"
	} else {
		return ok, "no"
	}

}
func (l *OrderUtilLogic) UpdateCoupon(order *cachemodel.Order, use bool) (bool, string) {
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
