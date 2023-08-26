package orderpay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"strconv"
	"strings"
	"time"
)

type OrderUtilLogic struct {
	logx.Logger
	ctx                   context.Context
	svcCtx                *svc.ServiceContext
	userphone             string
	ProductsMap           map[int64]*cachemodel.Product
	req                   *types.NewOrderRes
	ProductTinyList       []*types.ProductTiny
	coupon                *cachemodel.Coupon
	usercoupon            *cachemodel.UserCoupon
	UsedCouponStorInfo    map[int64]map[string]*types.CouponStoreInfo
	OrdersList            []*cachemodel.Order
	PayInit               *types.PayInit
	ul                    *UtilLogic
	ReallyUseCoupon       bool
	ProductQuantityInfoDB map[int64]map[string]*types.QuantityInfoDB
}

func NewOrderUtilLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderUtilLogic {
	return &OrderUtilLogic{
		Logger:             logx.WithContext(ctx),
		ctx:                ctx,
		svcCtx:             svcCtx,
		userphone:          ctx.Value("phone").(string),
		ul:                 NewUtilLogic(ctx, svcCtx),
		OrdersList:         make([]*cachemodel.Order, 0),
		UsedCouponStorInfo: make(map[int64]map[string]*types.CouponStoreInfo),
	}
}
func UpdatePayInfoIfFinished(svcCtx *svc.ServiceContext, PayInfo *cachemodel.PayInfo, OutTradeNo string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	if PayInfo == nil {
		PayInfo, _ = svcCtx.PayInfo.FindOneByOutTradeNo(context.Background(), OutTradeNo)
		if total, _, _ := IfFinished(PayInfo); total {
			svcCtx.PayInfo.UpdateAllPay(context.Background(), PayInfo.OutTradeNo)
		}
	}
	if total, _, _ := IfFinished(PayInfo); total {
		svcCtx.PayInfo.UpdateAllPay(context.Background(), PayInfo.OutTradeNo)
	}
}
func PayInfo2req(PayInfo *cachemodel.PayInfo, ContinuePayReq *types.ContinuePayRes) *types.NewOrderRes {
	ProductTinyList := make([]*types.ProductTiny, 0)
	json.Unmarshal([]byte(PayInfo.Pidlist), &ProductTinyList)
	NewOrderRes := &types.NewOrderRes{
		ProductTinyList: ProductTinyList,
		Address:         ContinuePayReq.Address,
		OrderNote:       ContinuePayReq.OrderNote,
		UsedCouponId:    ContinuePayReq.UsedCouponId,
		UsedCouponUUID:  ContinuePayReq.UsedCouponUUID,
		UseCouponFirst:  ContinuePayReq.UseCouponFirst,
		UseCashFirst:    ContinuePayReq.UseCashFirst,
	}
	return NewOrderRes
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
	if len(order.UsedCouponinfo) > 6 {
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
	PMcache, ok := l.svcCtx.LocalCache.Get(svc.ProductsMap)
	if ok {
		l.ProductsMap = PMcache.(map[int64]*cachemodel.Product)
	}

	OriginalAmount = int64(0)
	PromotionAmount = int64(0)
	l.ProductTinyList = ProductTinyList
	ok, ptlists := l.ProductTinyListChina()
	if !ok {
		return 0, 0
	}
	for _, ptlist := range ptlists {
		originalAmount, promotionAmount, _, _ := l.OriProPrice(ptlist)
		OriginalAmount = OriginalAmount + originalAmount
		PromotionAmount = PromotionAmount + promotionAmount
	}
	return OriginalAmount, PromotionAmount
}
func (l OrderUtilLogic) req2op(req *types.NewOrderRes) ([]*cachemodel.Order, *types.PayInit, bool) {
	//defer func() {
	//	if e := recover(); e != nil {
	//		return
	//	}
	//}()
	l.req = req
	PMcache, ok := l.svcCtx.LocalCache.Get(svc.ProductsMap)
	if ok {
		l.ProductsMap = PMcache.(map[int64]*cachemodel.Product)
	}
	ProductQuantityInfoDB, ok := l.svcCtx.LocalCache.Get(svc.ProductQuantityInfoDB)
	if ok {
		l.ProductQuantityInfoDB = ProductQuantityInfoDB.(map[int64]map[string]*types.QuantityInfoDB)
	}
	if l.req.UseCouponFirst {
		get, ok := l.svcCtx.LocalCache.Get(svc.CouponMapKey)

		if ok {
			l.coupon, ok = get.(map[int64]*cachemodel.Coupon)[l.req.UsedCouponId]
		}
		l.usercoupon, _ = l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, l.userphone)
		l.ReallyUseCoupon = l.CouponEffective()
	}
	l.ProductTinyList = req.ProductTinyList
	l.PayInit = &types.PayInit{}
	l.PayInit.Phone = l.userphone
	l.PayInit.OutTradeSn = RandStr(32)
	l.PayInit.NeedCashAccount = l.req.UseCashFirst
	l.PayInit.ProductTinyList = l.req.ProductTinyList
	return l.EndPayInit(l.OrderChina()), l.PayInit, true
}
func (l OrderUtilLogic) ConsumeInfo(ProductTinyList []*types.ProductTiny) {

}
func (l OrderUtilLogic) EndPayInit(OrdersList []*cachemodel.Order) []*cachemodel.Order {
	AllPromotionAmount := int64(0)
	for _, order := range OrdersList {
		AllPromotionAmount = AllPromotionAmount + order.PromotionAmount
	}
	_, AllCut := l.CountCouponPrice(AllPromotionAmount)
	AllCutReal := AllCut
	if len(l.UsedCouponStorInfo) > 0 {
		UsedCouponStorInfoByteList, _ := json.Marshal(l.UsedCouponStorInfo)
		UsedCouponStorInfoStr := string(UsedCouponStorInfoByteList)
		for i, _ := range OrdersList {
			part := float64(OrdersList[i].PromotionAmount) / float64(AllPromotionAmount)
			RealCut := int64(float64(AllCut) * part)
			AllCutReal = AllCutReal - RealCut
			OrdersList[i].ActualAmount = OrdersList[i].ActualAmount - RealCut
			OrdersList[i].UsedCouponinfo = UsedCouponStorInfoStr
		}
		OrdersList[0].ActualAmount = OrdersList[0].ActualAmount - AllCutReal
	}
	for _, order := range OrdersList {
		l.PayInit.TotleAmmount = l.PayInit.TotleAmmount + order.ActualAmount
	}
	return OrdersList
}
func (l OrderUtilLogic) ProductTinyListChina() (bool, [][]*types.ProductTiny) {
	//defer func() {
	//	if e := recover(); e != nil {
	//		return
	//	}
	//}()
	//for _, tiny := range l.ProductTinyList {
	//	ProductTinyMap, ok := l.MarketPlayerMap[l.ProductsMap[tiny.PId].MarketPlayerId]
	//	if ok {
	//		ProductTinyMap = l.PidListMapChina(ProductTinyMap, tiny)
	//	} else {
	//		ProductTinyMap = make(map[int64][][]*types.ProductTiny)
	//		ProductTinyMap = l.PidListMapChina(ProductTinyMap, tiny)
	//	}
	//	l.MarketPlayerMap[l.ProductsMap[tiny.PId].MarketPlayerId] = ProductTinyMap
	//}
	ProductTinyListForOrder := make([][]*types.ProductTiny, 0)
	for _, tiny := range l.ProductTinyList {
		for i := 0; i < int(tiny.Amount); i++ {
			ProductTinyListForOrder = append(ProductTinyListForOrder, []*types.ProductTiny{{tiny.PId, tiny.QuantityName, 1}})
		}
	}
	return true, ProductTinyListForOrder
}
func (l OrderUtilLogic) OrderChina() []*cachemodel.Order {
	//defer func() {
	//	if e := recover(); e != nil {
	//		return
	//	}
	//}()
	ok, ptlist := l.ProductTinyListChina()
	if !ok {
		return make([]*cachemodel.Order, 0)
	}
	return l.PidListMap2OrderMap(ptlist)
}
func (l OrderUtilLogic) PidListMap2OrderMap(ptlists [][]*types.ProductTiny) []*cachemodel.Order {
	orderlist := make([]*cachemodel.Order, 0)
	for _, ptlist := range ptlists {
		orderlist = append(orderlist, l.PidList2Order(ptlist))
	}
	//for MarketPlayerId, PidListMap := range l.MarketPlayerMap {
	//	for _, PidList := range PidListMap {
	//		orderlist = append(orderlist, l.PidList2Order(PidList, MarketPlayerId))
	//	}
	//}
	return orderlist
}
func (l OrderUtilLogic) GetPromotionPrice(Tiny *types.ProductTiny) int64 {
	if !l.ReallyUseCoupon {
		switch l.ProductsMap[Tiny.PId].Status {
		case 3, 4:
			return l.ProductQuantityInfoDB[Tiny.PId][Tiny.QuantityName].PromotionPrice - l.ProductQuantityInfoDB[Tiny.PId][Tiny.QuantityName].Cut
		}
	}
	return l.ProductQuantityInfoDB[Tiny.PId][Tiny.QuantityName].PromotionPrice
}
func (l OrderUtilLogic) OriProPrice(ProductTinyList []*types.ProductTiny) (OriginalAmount, PromotionAmount, ActualAmount int64, ProductInfoForSf string) {
	OriginalAmount = int64(0)
	PromotionAmount = int64(0)
	ProductInfoForSf = ""
	for _, tiny := range ProductTinyList {
		ProductInfoForSf = fmt.Sprintf("%s%s %s * %d %s", ProductInfoForSf, l.ProductsMap[tiny.PId].ProductCategoryName, tiny.QuantityName, tiny.Amount, "\n")
		OriginalAmount = OriginalAmount + l.ProductQuantityInfoDB[tiny.PId][tiny.QuantityName].OriginalPrice*tiny.Amount
		PromotionAmount = PromotionAmount + l.ProductQuantityInfoDB[tiny.PId][tiny.QuantityName].PromotionPrice*tiny.Amount
		ActualAmount = ActualAmount + l.GetPromotionPrice(tiny)*tiny.Amount
	}
	return OriginalAmount, PromotionAmount, ActualAmount, ProductInfoForSf
}
func (l OrderUtilLogic) GetOrderProductInfo(tiny *types.ProductTiny) (*types.OrderProductInfo, bool) {

	product := l.ProductsMap[tiny.PId]
	titleinfo := strings.Split(product.ProductTitle, "#")
	infoDB := l.ProductQuantityInfoDB[tiny.PId][tiny.QuantityName]
	OrderProductInfo := &types.OrderProductInfo{
		PId:             product.Pid,
		Amount:          tiny.Amount,
		PIdQuantity:     strconv.FormatInt(product.Pid, 10) + infoDB.Name,
		Picture:         product.Picture,
		ProductTitle:    titleinfo[0],
		ProductStandard: titleinfo[1],
		QuantityName:    infoDB.Name,
		PromotionPrice:  float64(infoDB.PromotionPrice) / 100,
		OriginalPrice:   float64(infoDB.OriginalPrice) / 100,
		IfCut:           getQuantityBool(infoDB.Cut),
		Cut:             float64(infoDB.Cut) / 100,
		SpecialPrice:    float64(infoDB.PromotionPrice-infoDB.Cut) / 100,
		Description:     product.Description,
		IfReserve:       getIfReserve(product.Status),
	}
	return OrderProductInfo, true
}
func (l OrderUtilLogic) GetOrderProductInfoList(ProductTinyList []*types.ProductTiny) ([]*types.OrderProductInfo, bool) {

	OrderProductInfoList := make([]*types.OrderProductInfo, 0)
	for _, tiny := range ProductTinyList {
		info, ok := l.GetOrderProductInfo(tiny)
		if ok {
			OrderProductInfoList = append(OrderProductInfoList, info)
		} else {
			return OrderProductInfoList, false
		}
	}
	return OrderProductInfoList, true
}
func (l OrderUtilLogic) PidList2Order(ProductTinyList []*types.ProductTiny) *cachemodel.Order {
	order := &cachemodel.Order{}
	order.OrderType = status2key(l.ProductsMap[ProductTinyList[0].PId].Status)
	order.MarketPlayerId = l.ProductsMap[ProductTinyList[0].PId].MarketPlayerId
	inittime, _ := time.Parse("2006-01-02 15:04:05", "2099-01-01 00:00:00")
	order.Phone = l.userphone
	order.OutTradeNo = l.PayInit.OutTradeSn
	order.OutRefundNo = RandStr(64)
	order.CreateOrderTime = time.Now()
	OrderProductInfo, ok := l.GetOrderProductInfoList(ProductTinyList)
	if ok {
		marshal, _ := json.Marshal(OrderProductInfo)
		order.Pidlist = string(marshal)
	}
	order.OriginalAmount, order.PromotionAmount, order.ActualAmount, order.ProductInfo = l.OriProPrice(ProductTinyList)
	order.CouponAmount = order.PromotionAmount - order.ActualAmount
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
	order.OrderSn = Getsha512(order.Phone + order.CreateOrderTime.String() + order.Pidlist + RandStr(64))
	order.LogId = time.Now().UnixNano()
	return order
}
func (l OrderUtilLogic) CountCouponPrice(PromotionPrice int64) (ActualAmount, Cut int64) {
	if l.coupon != nil && l.ReallyUseCoupon {
		if PromotionPrice >= l.coupon.MinPoint {
			return (PromotionPrice - l.coupon.Cut), l.coupon.Cut
		}
	}

	return PromotionPrice, 0
}
func (l *OrderUtilLogic) CouponEffective() bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	if l.coupon != nil && l.usercoupon != nil && (l.coupon.Cut != 0) {
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

func (l OrderUtilLogic) PidListMapChina(ProductTinyListMap map[int64][][]*types.ProductTiny, Tiny *types.ProductTiny) map[int64][][]*types.ProductTiny {
	key := status2key(l.ProductsMap[Tiny.PId].Status)
	if _, ok := ProductTinyListMap[key]; ok {
		for i := 0; i < int(Tiny.Amount); i++ {
			ProductTinyListMap[key] = append(ProductTinyListMap[key], []*types.ProductTiny{Tiny})
		}

	} else {
		ProductTinyListMap[key] = make([][]*types.ProductTiny, 0)
		for i := 0; i < int(Tiny.Amount); i++ {
			ProductTinyListMap[key] = append(ProductTinyListMap[key], []*types.ProductTiny{Tiny})
		}
	}
	return ProductTinyListMap
}
func (l *OrderUtilLogic) Updatecashaccount(phone, OrderSn, OutTradeNo string, CashAccountPayAmount, LogId int64, use bool) (bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	accphone := phone
	cashaccount, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, accphone)
	account, ok := cashfinish(CashAccountPayAmount, cashaccount, use)
	if ok {
		l.ul.Oplog("cash_account", OutTradeNo, "开始更新", LogId)
		l.svcCtx.CashAccount.Update(l.ctx, account)
		if use {
			l.svcCtx.PayInfo.UpdateCashPay(l.ctx, OutTradeNo)
			UpdatePayInfoIfFinished(l.svcCtx, nil, OutTradeNo)
			l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), OrderType: "消费", OrderSn: OutTradeNo, OrderDescribe: "购买店铺商品消费", Behavior: "消费", Phone: accphone, Balance: cashaccount.Balance, ChangeAmount: CashAccountPayAmount})

		} else {
			l.svcCtx.Order.RefundCash(l.ctx, OrderSn)
			l.svcCtx.PayInfo.UpdateCashReject(l.ctx, CashAccountPayAmount, OutTradeNo)
			l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), OrderType: "退款", OrderSn: OutTradeNo, OrderDescribe: "店铺商品退款", Behavior: "退款", Phone: accphone, Balance: cashaccount.Balance, ChangeAmount: CashAccountPayAmount})
		}
		l.ul.Oplog("cash_account", OutTradeNo, "结束更新", LogId)

		return ok, "yes"
	} else {
		return ok, "no"
	}

}
func (l *OrderUtilLogic) UpdateCoupon(order *cachemodel.Order, use bool) (bool, string) {
	//defer func() {
	//	if e := recover(); e != nil {
	//		return
	//	}
	//}()
	usercoupon, _ := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, l.userphone)
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
func OrderNeedChange(order *cachemodel.Order) bool {
	if order.OrderStatus == 0 || order.OrderStatus == 6 {
		return true
	} else {
		return false
	}
}
