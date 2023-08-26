package orderpay

import (
	"context"
	"math/rand"
	"oa_final/cachemodel"
	"strconv"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CashrechargeLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	lid             int64
	userphone       string
	useropenid      string
	WeChatUtilLogic *WeChatUtilLogic
}

func NewCashrechargeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CashrechargeLogic {
	return &CashrechargeLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		userphone:       ctx.Value("phone").(string),
		useropenid:      ctx.Value("openid").(string),
		WeChatUtilLogic: NewWeChatUtilLogic(ctx, svcCtx),
	}
}

func (l *CashrechargeLogic) Cashrecharge(req *types.CashRechargeRes) (resp *types.CashRechargeResp, err error) {
	l.lid = time.Now().UnixNano() + int64(rand.Intn(1024))
	get, ok := l.svcCtx.LocalCache.Get(svc.RechargeProductKey)
	if !ok {
		return &types.CashRechargeResp{Code: "4004", Msg: "数据库失效"}, nil
	}
	rcpmap := get.(map[int64]*cachemodel.RechargeProduct)
	rproduct, ok := rcpmap[req.Rpid]
	if !ok {
		return &types.CashRechargeResp{Code: "4004", Msg: "无此rpid"}, nil
	}
	orderdb := l.order2db(rproduct)
	l.svcCtx.RechargeOrder.Insert(l.ctx, orderdb)
	order, _ := l.svcCtx.RechargeOrder.FindOneByOrderSn(l.ctx, orderdb.OrderSn)
	if order == nil {
		return &types.CashRechargeResp{Code: "4004", Msg: "数据库失效"}, nil
	}
	weixinpayinit := l.WeChatUtilLogic.Weixinpayinit(order.OutTradeNo, order.WexinPayAmount, UseRecallDir("/payrecall/coupontellmeso"))
	cashrecharge := &types.CashRechargeRp{RechargeOrderInfo: db2info(order), WeiXinPayMsg: weixinpayinit}
	return &types.CashRechargeResp{Code: "10000", Msg: "success", Data: cashrecharge}, nil
}

func db2info(order *cachemodel.RechargeOrder) *types.RechargeOrderInfo {
	info := &types.RechargeOrderInfo{}
	info.Amount = order.Amount
	info.GiftAmount = order.GiftAmount
	info.WexinPayAmount = order.Amount
	info.Rpid = order.Rpid
	info.Phone = order.Phone
	info.OrderSn = order.OrderSn
	info.OutTradeSn = order.OutTradeNo
	info.TransactionId = order.TransactionId
	info.PaymentTime = order.PaymentTime.Format("2006-01-02 15:04:05")
	info.CreateOrderTime = order.CreateOrderTime.Format("2006-01-02 15:04:05")
	return info
}
func (l *CashrechargeLogic) order2db(rproduct *cachemodel.RechargeProduct) *cachemodel.RechargeOrder {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "2099-01-01 00:00:00")
	db := &cachemodel.RechargeOrder{}
	db.LogId = l.lid
	db.OutTradeNo = RandStr(32)
	db.Rpid = rproduct.Rpid
	db.Phone = l.userphone
	db.CreateOrderTime = time.Now()
	db.PaymentTime = inittime
	db.OrderStatus = 0
	db.WexinPayAmount = rproduct.Price
	db.GiftAmount = rproduct.GiftAmount
	db.Amount = rproduct.Price
	db.OrderSn = Getsha512(db.Phone + db.CreateOrderTime.String() + strconv.FormatInt(db.Rpid, 10))
	return db
}
