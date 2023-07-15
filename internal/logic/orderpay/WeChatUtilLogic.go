package orderpay

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"math/rand"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

type WeChatUtilLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	userphone   string
	useropenid  string
	tellmesodir string
}

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type WeChatPayOpt func(*WeChatUtilLogic)

func NewWeChatUtilLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WeChatUtilLogic {
	return &WeChatUtilLogic{
		Logger:      logx.WithContext(ctx),
		ctx:         ctx,
		svcCtx:      svcCtx,
		userphone:   ctx.Value("phone").(string),
		useropenid:  ctx.Value("openid").(string),
		tellmesodir: svcCtx.Config.ServerInfo.Url + "/payrecall/tellmeso",
	}
}
func UseRecallDir(dir string) WeChatPayOpt {
	return func(l *WeChatUtilLogic) {
		l.tellmesodir = l.svcCtx.Config.ServerInfo.Url + dir
	}
}
func (l *WeChatUtilLogic) CheckWeiXinPayFinished(OutTradeNo string) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	no2payment, result, _ := jssvc.QueryOrderByOutTradeNo(l.ctx, jsapi.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(OutTradeNo),
		Mchid:      core.String(l.svcCtx.Config.WxConf.MchID)})
	defer result.Response.Body.Close()
	if *no2payment.TradeState != "SUCCESS" {
		return true
	}
	return false
}
func (l *WeChatUtilLogic) CancelOrder(order *cachemodel.Order) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()

	payInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, order.OutTradeNo)
	if payInfo.WexinRefundAmount+order.WexinPayAmount > (payInfo.WexinPayAmount + 5) {
		return false
	}
	service := refunddomestic.RefundsApiService{Client: l.svcCtx.Client}
	create, result, err := service.Create(l.ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(order.OutTradeNo),
		OutRefundNo: core.String(order.OutRefundNo),
		Amount: &refunddomestic.AmountReq{Currency: core.String("CNY"),
			Refund: core.Int64(order.WexinPayAmount),
			Total:  core.Int64(payInfo.WexinPayAmount)},
	})
	defer result.Response.Body.Close()
	if err != nil {
		log.Printf("call Create err:%s", err)
		return false
	} else {
		log.Printf("status=%d resp=%s", result.Response.StatusCode, result.Response, create.String())
	}
	l.svcCtx.Order.RefundWeChat(l.ctx, order.OrderSn)
	l.svcCtx.PayInfo.UpdateWeixinReject(l.ctx, order.WexinPayAmount, order.OutTradeNo)

	//l.svcCtx.RefundInfo.Insert(l.ctx, &cachemodel.RefundInfo{Phone: l.userphone, OutTradeNo: payInfo.OutTradeNo, OutRefundNo: OutRefundNo, RefundAmount: order.WexinPayAmount, WexinRefundTime: time.Now()})
	return true
}
func (l *WeChatUtilLogic) IfCancelOrderSuccess(order *cachemodel.Order) bool {
	service := refunddomestic.RefundsApiService{Client: l.svcCtx.Client}
	no, result, err := service.QueryByOutRefundNo(l.ctx,
		refunddomestic.QueryByOutRefundNoRequest{
			OutRefundNo: core.String(order.OutRefundNo),
		})
	defer result.Response.Body.Close()
	if err != nil {
		log.Printf("call QueryByOutRefundNo err:%s", err)
		return false

	} else {
		log.Printf("status=%d resp=%s", result.Response.StatusCode, result.Response)
	}
	if *no.Status == "SUCCESS" {
		l.svcCtx.Order.RefundWeChat(l.ctx, order.OrderSn)
		l.svcCtx.PayInfo.UpdateWeixinReject(l.ctx, order.WexinPayAmount, order.OutTradeNo)
		return true
	}
	return false
}
func (l *WeChatUtilLogic) Weixinpayinit(OutTradeNo string, ammount int64, options ...func(logic *WeChatUtilLogic)) *types.WeChatPayMsg {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	for _, option := range options {
		option(l)
	}
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	// 得到prepay_id，以及调起支付所需的参数和签名
	payment, result, err := jssvc.PrepayWithRequestPayment(l.ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(l.svcCtx.Config.WxConf.AppId),
			Mchid:       core.String(l.svcCtx.Config.WxConf.MchID),
			Description: core.String("沾还是不沾芥末，这是一个问题"),
			OutTradeNo:  core.String(OutTradeNo),
			Attach:      core.String(randStr(16)),
			NotifyUrl:   core.String(l.tellmesodir),
			Amount: &jsapi.Amount{
				Total: core.Int64(ammount),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(l.ctx.Value("openid").(string)),
			},
		},
	)
	defer result.Response.Body.Close()
	if err == nil {
		log.Println(payment, result)
	} else {
		log.Println(err)
	}
	// 用于返回给前端调起支付的变量与签名串生成器
	timestampsec := *payment.TimeStamp
	nonceStr := *payment.NonceStr
	packagestr := *payment.Package
	paySign := *payment.PaySign
	signType := *payment.SignType
	return &types.WeChatPayMsg{PaySign: paySign, NonceStr: nonceStr, TimeStamp: timestampsec, Package: packagestr, SignType: signType}
}
func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
func (l *WeChatUtilLogic) CheckWeiXinPay(OutTradeNo string) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	no2payment, result, _ := jssvc.QueryOrderByOutTradeNo(l.ctx, jsapi.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(OutTradeNo),
		Mchid:      core.String(l.svcCtx.Config.WxConf.MchID)})
	defer result.Response.Body.Close()
	fmt.Println(*no2payment.TradeState, "%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
	if *no2payment.TradeState == "SUCCESS" {
		l.svcCtx.PayInfo.UpdateWeixinPay(context.Background(), OutTradeNo, *no2payment.TransactionId)
		return true
	}
	return false
}
