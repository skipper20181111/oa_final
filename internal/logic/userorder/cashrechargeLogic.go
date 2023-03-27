package userorder

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"log"
	"math/rand"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type CashrechargeLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
	lid       int64
}

func NewCashrechargeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CashrechargeLogic {
	return &CashrechargeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CashrechargeLogic) Cashrecharge(req *types.CashRechargeRes) (resp *types.CashRechargeResp, err error) {
	l.userphone = l.ctx.Value("phone").(string)
	useropenid := l.ctx.Value("openid").(string)
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
	l.oplog("现金账户充值", order.OrderSn, "开始更新", orderdb.LogId)
	if order == nil {
		return &types.CashRechargeResp{Code: "4004", Msg: "数据库失效"}, nil
	}
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	// 得到prepay_id，以及调起支付所需的参数和签名
	payment, result, err := jssvc.PrepayWithRequestPayment(l.ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(l.svcCtx.Config.WxConf.AppId),
			Mchid:       core.String(l.svcCtx.Config.WxConf.MchID),
			Description: core.String("不沾芥末没灵魂"),
			OutTradeNo:  core.String(order.OutTradeNo),
			Attach:      core.String(randStr(16)),
			NotifyUrl:   core.String(l.svcCtx.Config.ServerInfo.Url + "/payrecall/coupontellmeso"),
			Amount: &jsapi.Amount{
				Total: core.Int64(order.WexinPayAmount),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(useropenid),
			},
		},
	)
	defer result.Response.Body.Close()
	if err == nil {
		log.Println(payment, result)
	} else {
		log.Println(err)
		return &types.CashRechargeResp{Code: "4004", Msg: err.Error()}, nil
	}
	// 用于返回给前端调起支付的变量与签名串生成器
	timestampsec := *payment.TimeStamp
	nonceStr := *payment.NonceStr
	packagestr := *payment.Package
	paySign := *payment.PaySign
	signType := *payment.SignType
	cashrecharge := &types.CashRechargeRp{RechargeOrderInfo: db2info(order), TimeStamp: timestampsec, NonceStr: nonceStr, Package: packagestr, SignType: signType, PaySign: paySign}
	return &types.CashRechargeResp{Code: "10000", Msg: "success", Data: cashrecharge}, nil
}
func db2info(order *cachemodel.RechargeOrder) *types.RechargeOrderInfo {
	info := &types.RechargeOrderInfo{}
	info.Amount = order.Amount
	info.GiftAmount = order.Amount
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
func (l *CashrechargeLogic) oplog(tablename, event, describe string, lid int64) error {
	aol := &cachemodel.AccountOperateLog{Phone: l.ctx.Value("phone").(string), TableName: tablename, Event: event, Describe: describe, Timestamp: time.Now(), Lid: lid}
	_, err := l.svcCtx.AccountOperateLog.Insert(l.ctx, aol)
	return err
}
func (l *CashrechargeLogic) order2db(rproduct *cachemodel.RechargeProduct) *cachemodel.RechargeOrder {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	db := &cachemodel.RechargeOrder{}
	db.LogId = l.lid
	db.OutTradeNo = randStr(32)
	db.Rpid = rproduct.Rpid
	db.Phone = l.userphone
	db.CreateOrderTime = time.Now()
	db.PaymentTime = inittime
	db.OrderStatus = 0
	db.WexinPayAmount = rproduct.Price * 100
	db.GiftAmount = rproduct.GiftAmount * 100
	db.Amount = rproduct.Price * 100
	db.OrderSn = getsha512(db.Phone + db.CreateOrderTime.String() + strconv.FormatInt(db.Rpid, 10))
	return db
}
