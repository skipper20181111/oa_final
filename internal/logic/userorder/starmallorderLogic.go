package userorder

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

	get, ok := l.svcCtx.LocalCache.Get(svc.StarMallMap)
	if !ok {
		return &types.StarMallOrderResp{Code: "4004", Msg: "缓存失效"}, nil
	}
	orderinfo := &types.OrderInfo{}
	StarMallMap := get.(map[int64]*cachemodel.StarmallLonglist)
	_, ok = StarMallMap[req.Pid]
	if !ok {
		return &types.StarMallOrderResp{Code: "10000", Msg: "无此商品", Data: orderinfo}, nil
	}
	db := starreq2db(req, l.ctx.Value("phone").(string))
	return &types.StarMallOrderResp{Code: "10000", Msg: "success", Data: OrderDb2info(db)}, nil
}
func starreq2db(req *types.StarMallOrderRes, phone string) *cachemodel.UserOrder {

	inittime, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	order := &cachemodel.UserOrder{}
	order.Phone = phone
	order.PointSorder = 1
	order.FinishWeixinpay = 0
	order.FinishAccountpay = 0
	order.CreateOrderTime = time.Now()
	order.OutTradeNo = randStr(32)
	ProductTinyList := [1]*types.ProductTiny{&types.ProductTiny{PId: req.Pid, Amount: 1}}
	marshal, err := json.Marshal(ProductTinyList)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	order.Pidlist = string(marshal)

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
	return order

}
