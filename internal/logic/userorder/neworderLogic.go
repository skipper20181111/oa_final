package userorder

import (
	"context"
	"crypto"
	CRand "crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"math/rand"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"strconv"
	"time"
)

type NeworderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNeworderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NeworderLogic {
	return &NeworderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NeworderLogic) Neworder(req *types.NewOrderRes) (resp *types.NewOrderResp, err error) {
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
			NotifyUrl:   core.String("https://www.weixin.qq.com/wxpay/pay.php"),
			Amount: &jsapi.Amount{
				Total: core.Int64(req.Money),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(req.OpenId),
			},
		},
	)
	if err == nil {
		log.Println(payment, result)
	} else {
		log.Println(err)
	}
	// 用于返回给前端调起支付的变量与签名串生成器
	timestampsec := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := randStr(32)
	packagestr := *payment.Package
	src := l.svcCtx.Config.WxConf.AppId + "\n" + timestampsec + "\n" + nonceStr + "\n" + packagestr + "\n"
	sign, err := Sign(l.svcCtx.MchPrivateKey, src)
	paySign := base64.StdEncoding.EncodeToString(sign)
	signType := "RSA"
	neworderrp := types.NewOrderRp{TimeStamp: timestampsec, NonceStr: nonceStr, Package: packagestr, SignType: signType, PaySign: paySign}
	return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &neworderrp}, nil
}

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
func Sign(key *rsa.PrivateKey, src string) (sign []byte, err error) {
	h := crypto.SHA256
	hn := h.New()
	hn.Write([]byte(src))
	sum := hn.Sum(nil)
	return rsa.SignPKCS1v15(CRand.Reader, key, h, sum)
	//return rsa.SignPSS(rand.Reader, key, h, sum, nil)
}
