package orderpay

import (
	"context"
	"encoding/json"
	"fmt"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StarmallorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStarmallorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StarmallorderLogic {
	return &StarmallorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StarmallorderLogic) Starmallorder(req *types.StarMallOrderRes) (resp *types.StarMallOrderResp, err error) {
	phone := l.ctx.Value("phone").(string)
	get, ok := l.svcCtx.LocalCache.Get(svc.StarMallMap)
	if !ok {
		return &types.StarMallOrderResp{Code: "4004", Msg: "缓存失效"}, nil
	}
	orderinfo := &types.OrderInfo{}
	StarMallMap := get.(map[int64]*cachemodel.StarmallLonglist)
	_, ok = StarMallMap[req.Pid]
	if !ok || StarMallMap[req.Pid].ExchangePoints == 0 {
		return &types.StarMallOrderResp{Code: "10000", Msg: "无此商品或此商品不可兑换", Data: orderinfo}, nil
	}
	cache, err := l.svcCtx.UserPoints.FindOneByPhone(l.ctx, phone)
	if cache != nil && cache.AvailablePoints > StarMallMap[req.Pid].ExchangePoints {
		db := starreq2db(req, phone, StarMallMap[req.Pid].ExchangePoints)
		cache.AvailablePoints = cache.AvailablePoints - StarMallMap[req.Pid].ExchangePoints
		l.svcCtx.UserPoints.Update(l.ctx, cache)
		l.svcCtx.Order.Insert(l.ctx, db)
		l.insertstarTransaction(db)
		l.svcCtx.PointLog.Insert(l.ctx, &cachemodel.PointLog{Date: time.Now(), OrderType: "兑换商品", OrderSn: db.OrderSn, OrderDescribe: "臻星商城兑换商品", Behavior: "兑换", Phone: db.Phone, Balance: cache.AvailablePoints, ChangeAmount: StarMallMap[req.Pid].ExchangePoints})

		return &types.StarMallOrderResp{Code: "10000", Msg: "success", Data: OrderDb2info(db)}, nil
	} else {
		return &types.StarMallOrderResp{Code: "10000", Msg: "积分不足", Data: orderinfo}, nil
	}
}
func (l *StarmallorderLogic) insertstarTransaction(order *cachemodel.Order) {
	transaction := &cachemodel.TransactionInfo{}
	transaction.Phone = order.Phone
	transaction.OrderSn = order.OrderSn
	transaction.OutTradeNo = order.OutTradeNo
	transaction.TransactionType = "积分兑换商品"
	transaction.CreateOrderTime = order.CreateOrderTime
	transaction.FinishWeixinpay = 1
	transaction.FinishAccountpay = 1
	transaction.Status = 1
	transaction.LogId = order.LogId
	transaction.WexinPaymentTime = order.CreateOrderTime
	transaction.CashAccountPaymentTime = order.CreateOrderTime
	l.svcCtx.TransactionInfo.Insert(l.ctx, transaction)
}
func starreq2db(req *types.StarMallOrderRes, phone string, pointamount int64) *cachemodel.Order {

	inittime, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	order := &cachemodel.Order{}
	order.Phone = phone
	order.PointsOrder = 1
	order.FinishWeixinpay = 0
	order.FinishAccountpay = 0
	//order.PointAmount = pointamount
	order.CreateOrderTime = time.Now()
	order.OutTradeNo = randStr(32)
	ProductTinyList := [1]*types.ProductTiny{&types.ProductTiny{PId: req.Pid, Amount: 1}}
	marshal, err := json.Marshal(ProductTinyList)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	order.Pidlist = string(marshal)

	order.FreightAmount = 4000
	order.OrderStatus = 1
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
	order.PaymentTime = order.CreateOrderTime
	order.DeliveryTime = inittime
	order.ReceiveTime = inittime
	order.CloseTime = inittime
	order.OrderSn = getsha512(order.Phone + order.CreateOrderTime.String() + order.Pidlist + order.Address)
	order.LogId = time.Now().UnixMicro()
	return order

}
