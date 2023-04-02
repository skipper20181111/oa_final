package userorder

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"log"
	"math/rand"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type NeworderLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	cashaccount *cachemodel.CashAccount
	userorder   *cachemodel.UserOrder
	usecash     bool
	usecoupon   bool
	usepoint    bool
	userphone   string
}

func NewNeworderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NeworderLogic {
	return &NeworderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NeworderLogic) Neworder(req *types.NewOrderRes) (resp *types.NewOrderResp, err error) {
	lid := time.Now().UnixNano() + int64(rand.Intn(1024))
	UseAccount := false
	if len(req.ProductTinyList) == 0 {
		return &types.NewOrderResp{Code: "4004", Msg: "无商品，订单金额为0", Data: &types.NewOrderRp{}}, nil
	}
	PMcache, ok := l.svcCtx.LocalCache.Get(svc.ProductsMap)
	if !ok {
		return &types.NewOrderResp{Code: "4004", Msg: "服务器查找商品列表失败"}, nil
	}
	productsMap := PMcache.(map[int64]*types.ProductInfo)
	lu := NewLogic(l.ctx, l.svcCtx)
	order := lu.Order2db(req, productsMap, UseCache(false))
	order.LogId = lid
	lu.oplog("付款啊", order.OrderSn, "开始更新", lid)
	l.svcCtx.UserOrder.Insert(l.ctx, order)
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, order.OrderSn)
	if sn2order == nil {
		fmt.Println(err.Error())
		return &types.NewOrderResp{Code: "4004", Msg: "数据库失效"}, nil
	}
	l.userorder = sn2order
	orderinfo := OrderDb2info(sn2order)
	if lu.usecoupon || lu.usecash || lu.usepoint {
		UseAccount = true
	}
	money := sn2order.WexinPayAmount
	if money == 0 {
		return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &types.NewOrderRp{OrderInfo: orderinfo, UseWechatPay: false, UseAccount: UseAccount}}, nil
	}
	// 此处开始生成订单
	lu.oplog("微信支付啊", l.userorder.OrderSn, "开始更新", lid)
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	// 得到prepay_id，以及调起支付所需的参数和签名
	payment, result, err := jssvc.PrepayWithRequestPayment(l.ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(l.svcCtx.Config.WxConf.AppId),
			Mchid:       core.String(l.svcCtx.Config.WxConf.MchID),
			Description: core.String("沾还是不沾芥末，这是一个问题"),
			OutTradeNo:  core.String(orderinfo.OutTradeNo),
			Attach:      core.String(randStr(16)),
			NotifyUrl:   core.String(l.svcCtx.Config.ServerInfo.Url + "/payrecall/tellmeso"),
			Amount: &jsapi.Amount{
				Total: core.Int64(money),
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
		return &types.NewOrderResp{Code: "4004", Msg: err.Error()}, nil
	}
	// 用于返回给前端调起支付的变量与签名串生成器
	timestampsec := *payment.TimeStamp
	nonceStr := *payment.NonceStr
	packagestr := *payment.Package
	paySign := *payment.PaySign
	signType := *payment.SignType
	neworderrp := types.NewOrderRp{OrderInfo: orderinfo, UseAccount: UseAccount, UseWechatPay: true, TimeStamp: timestampsec, NonceStr: nonceStr, Package: packagestr, SignType: signType, PaySign: paySign}
	return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &neworderrp}, nil

}
