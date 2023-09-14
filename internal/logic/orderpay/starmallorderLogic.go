package orderpay

import (
	"context"
	"encoding/json"
	"fmt"
	"oa_final/cachemodel"
	"strconv"
	"strings"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StarmallorderLogic struct {
	logx.Logger
	ctx                   context.Context
	svcCtx                *svc.ServiceContext
	ProductsMap           map[int64]*cachemodel.Product
	ProductQuantityInfoDB map[int64]map[string]*types.QuantityInfoDB
	StarMallMap           map[int64]*cachemodel.StarmallLonglist
	req                   *types.StarMallOrderRes
	phone                 string
	u                     *UtilLogic
}

func NewStarmallorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StarmallorderLogic {
	return &StarmallorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		phone:  ctx.Value("phone").(string),
		u:      NewUtilLogic(ctx, svcCtx),
	}
}

func (l *StarmallorderLogic) Starmallorder(req *types.StarMallOrderRes) (resp *types.StarMallOrderResp, err error) {
	get, ok := l.svcCtx.LocalCache.Get(svc.StarMallMap)
	if ok {
		l.StarMallMap = get.(map[int64]*cachemodel.StarmallLonglist)
	}
	PMcache, ok := l.svcCtx.LocalCache.Get(svc.ProductsMap)
	if ok {
		l.ProductsMap = PMcache.(map[int64]*cachemodel.Product)
	}
	ProductQuantityInfoDB, ok := l.svcCtx.LocalCache.Get(svc.ProductQuantityInfoDB)
	if ok {
		l.ProductQuantityInfoDB = ProductQuantityInfoDB.(map[int64]map[string]*types.QuantityInfoDB)
	}
	orderinfo := &types.OrderInfo{}
	l.req = req
	order, ok, msg := l.InsertStarDb()
	if ok {
		return &types.StarMallOrderResp{Code: "10000", Msg: "success", Data: l.u.OrderDb2info(order)}, nil
	}
	return &types.StarMallOrderResp{Code: "10000", Msg: msg, Data: orderinfo}, nil
}
func (l *StarmallorderLogic) insertstarTransaction(order *cachemodel.Order) {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "2099-01-01 00:00:00")
	PayInfo := &cachemodel.PayInfo{}
	PayInfo.OutTradeNo = order.OutTradeNo
	PayInfo.Phone = order.Phone
	PayInfo.TransactionType = "StarMall"
	PayInfo.Status = 1
	PayInfo.CreateOrderTime = time.Now()
	PayInfo.CashAccountPaymentTime = PayInfo.CreateOrderTime
	PayInfo.WexinPaymentTime = PayInfo.CreateOrderTime
	PayInfo.LogId = PayInfo.CreateOrderTime.UnixNano()

	PayInfo.WexinDeliveryTime = inittime
	l.svcCtx.PayInfo.Insert(l.ctx, PayInfo)
}
func (l StarmallorderLogic) GetOrder(Starmall *cachemodel.StarmallLonglist, Product *cachemodel.Product, QuantityInfo *types.QuantityInfoDB) *cachemodel.Order {
	order := &cachemodel.Order{}
	order.OrderType = 0
	order.PointsOrder = 1
	order.OrderStatus = 1001
	inittime, _ := time.Parse("2006-01-02 15:04:05", "2099-01-01 00:00:00")
	order.Phone = l.phone
	order.OutTradeNo = RandStr(64)
	order.OutRefundNo = RandStr(64)
	order.CreateOrderTime = time.Now()
	order.PointsAmount = Starmall.ExchangePoints
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
	order.WexinDeliveryTime = inittime
	order.OrderSn = Getsha512(order.Phone + order.CreateOrderTime.String() + order.Pidlist + RandStr(64))
	order.LogId = time.Now().UnixNano()
	order.ProductInfo = fmt.Sprintf("%s %s * %d %s", Product.ProductCategoryName, QuantityInfo.Name, 1, "\n")
	OrderProductInfo, ok := l.GetOrderProductInfo(Product, QuantityInfo)
	if ok {
		marshal, _ := json.Marshal(OrderProductInfo)
		order.Pidlist = string(marshal)
	}
	l.svcCtx.Order.Insert(l.ctx, order)
	return order
}
func (l StarmallorderLogic) GetOrderProductInfo(Product *cachemodel.Product, QuantityInfo *types.QuantityInfoDB) ([]*types.OrderProductInfo, bool) {
	titleinfo := strings.Split(Product.ProductTitle, "#")
	OrderProductInfo := &types.OrderProductInfo{
		PId:             Product.Pid,
		Amount:          1,
		PIdQuantity:     strconv.FormatInt(Product.Pid, 10) + QuantityInfo.Name,
		Picture:         Product.Picture,
		ProductTitle:    titleinfo[0],
		ProductStandard: titleinfo[1],
		QuantityName:    QuantityInfo.Name,
		PromotionPrice:  float64(QuantityInfo.PromotionPrice) / 100,
		OriginalPrice:   float64(QuantityInfo.OriginalPrice) / 100,
		IfCut:           getQuantityBool(QuantityInfo.Cut),
		Cut:             float64(QuantityInfo.Cut) / 100,
		SpecialPrice:    float64(QuantityInfo.PromotionPrice-QuantityInfo.Cut) / 100,
		Description:     Product.Description,
		IfReserve:       getIfReserve(Product.Status),
	}
	return []*types.OrderProductInfo{OrderProductInfo}, true
}

func (l *StarmallorderLogic) InsertStarDb() (*cachemodel.Order, bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	Starmall := l.StarMallMap[l.req.Pid]
	Product := l.ProductsMap[l.req.Pid]
	QuantityInfo := l.ProductQuantityInfoDB[l.req.Pid][Starmall.QuantityName]
	UserPoints, _ := l.svcCtx.UserPoints.FindOneByPhone(l.ctx, l.phone)
	if UserPoints == nil || UserPoints.AvailablePoints < Starmall.ExchangePoints {
		return nil, false, "积分不足"
	} else {
		UserPoints.AvailablePoints = UserPoints.AvailablePoints - Starmall.ExchangePoints
		l.svcCtx.UserPoints.Update(l.ctx, UserPoints)
	}
	order := l.GetOrder(Starmall, Product, QuantityInfo)
	l.insertstarTransaction(order)
	l.svcCtx.PointLog.Insert(l.ctx, &cachemodel.PointLog{Date: time.Now(), OrderType: "兑换商品", OrderSn: order.OrderSn, OrderDescribe: "臻星商城兑换商品", Behavior: "兑换", Phone: l.phone, Balance: UserPoints.AvailablePoints, ChangeAmount: Starmall.ExchangePoints})
	return order, true, "success"
}
