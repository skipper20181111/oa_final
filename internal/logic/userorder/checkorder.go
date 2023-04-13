package userorder

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

type CheckOrderLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	userphone       string
	useropenid      string
	WeChatUtilLogic *WeChatUtilLogic
}

func NewCheckOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckOrderLogic {
	return &CheckOrderLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		userphone:       ctx.Value("phone").(string),
		useropenid:      ctx.Value("openid").(string),
		WeChatUtilLogic: NewWeChatUtilLogic(ctx, svcCtx),
	}
}

func (l *CheckOrderLogic) Payorder(req *types.TransactionInit) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
}
func (l *CheckOrderLogic) CheckWeiXinPay(OutTradeNo string) bool {
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
func (l *CheckOrderLogic) CheckWeiXinReject(Order *cachemodel.UserOrder) bool {
	return l.WeChatUtilLogic.IfCancelOrderSuccess(Order)
}
