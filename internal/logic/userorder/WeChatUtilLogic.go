package userorder

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
)

type WeChatUtilLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	userphone  string
	useropenid string
}

func NewWeChatUtilLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WeChatUtilLogic {
	return &WeChatUtilLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
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
func (l *WeChatUtilLogic) CancelOrder(order *cachemodel.UserOrder) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	service := refunddomestic.RefundsApiService{Client: l.svcCtx.Client}
	create, result, err := service.Create(l.ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(order.OutTradeNo),
		OutRefundNo: core.String(order.OutTradeNo),
		Amount: &refunddomestic.AmountReq{Currency: core.String("CNY"),
			Refund: core.Int64(order.WexinPayAmount),
			Total:  core.Int64(order.WexinPayAmount)},
	})
	defer result.Response.Body.Close()
	if err != nil {
		log.Printf("call Create err:%s", err)
		return false
	} else {
		log.Printf("status=%d resp=%s", result.Response.StatusCode, result.Response, create.String())
	}
	l.svcCtx.TransactionInfo.UpdateWeixinReject(l.ctx, order.OrderSn)
	return true
}
func (l *WeChatUtilLogic) IfCancelOrderSuccess(order *cachemodel.UserOrder) bool {
	service := refunddomestic.RefundsApiService{Client: l.svcCtx.Client}
	no, result, err := service.QueryByOutRefundNo(l.ctx,
		refunddomestic.QueryByOutRefundNoRequest{
			OutRefundNo: core.String(order.OutTradeNo),
		})
	defer result.Response.Body.Close()
	if err != nil {
		log.Printf("call QueryByOutRefundNo err:%s", err)
		return false

	} else {
		log.Printf("status=%d resp=%s", result.Response.StatusCode, result.Response)
	}
	if *no.Status == "SUCCESS" {
		return true
	}
	return false
}
