package userorder

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"log"
	"math/rand"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PaybillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaybillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaybillLogic {
	return &PaybillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PaybillLogic) Paybill(req *types.BillPayRes) (resp *types.BillPayResp, err error) {
	useropenid := l.ctx.Value("openid").(string)
	rand.Seed(time.Now().Unix())
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	// 得到prepay_id，以及调起支付所需的参数和签名
	payment, result, err := jssvc.PrepayWithRequestPayment(l.ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(l.svcCtx.Config.WxConf.AppId),
			Mchid:       core.String(l.svcCtx.Config.WxConf.MchID),
			Description: core.String("Image形象店-深圳腾大-QQ公仔"),
			OutTradeNo:  core.String(randStr(32)),
			Attach:      core.String("自定义数据说明" + randStr(5)),
			NotifyUrl:   core.String(l.svcCtx.Config.ServerInfo.Url + "/payrecall/tellmeso"),
			Amount: &jsapi.Amount{
				Total: core.Int64(req.Money),
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
		return &types.BillPayResp{Code: "4004", Msg: err.Error()}, nil
	}
	// 用于返回给前端调起支付的变量与签名串生成器
	timestampsec := *payment.TimeStamp
	nonceStr := *payment.NonceStr
	packagestr := *payment.Package
	paySign := *payment.PaySign
	signType := *payment.SignType
	billpayrp := types.BillPayRp{TimeStamp: timestampsec, NonceStr: nonceStr, Package: packagestr, SignType: signType, PaySign: paySign}
	return &types.BillPayResp{Code: "10000", Msg: "success", Data: &billpayrp}, nil

	return
}
