package userorder

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"time"
)

type PayLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	userphone       string
	req             *types.TransactionInit
	transantioninfo *cachemodel.TransactionInfo
	weixinpayinit   *types.WeiXinPayMsg
}

func NewPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayLogic {
	return &PayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayLogic) Payorder(req *types.TransactionInit) (resp *types.PayMsg, success bool) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	l.req = req
	l.transactioninfoinit()
	l.transactionend()
	resp = l.db2resp()
	sn, _ := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, req.OrderSn)
	if sn != nil {
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		l.svcCtx.TransactionInfo.Update(l.ctx, l.transantioninfo)
	} else {
		fmt.Println("(((((((((((((((((((((((((((((@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@)))))))))))))))))))))))))))))")
		l.svcCtx.TransactionInfo.Insert(l.ctx, l.transantioninfo)
	}
	return resp, true

}
func IfFinished(info *cachemodel.TransactionInfo) bool {
	cash := false
	weixin := false
	if info.NeedCashAccount == 1 && info.FinishAccountpay == 1 {
		cash = true
	}
	if info.NeedCashAccount != 1 {
		cash = true
	}
	if info.WexinPayAmount > 0 && info.FinishWeixinpay == 1 {
		weixin = true
	}
	if info.WexinPayAmount <= 0 {
		weixin = true
	}
	return weixin && cash
}
func (l *PayLogic) db2resp() *types.PayMsg {
	resp := &types.PayMsg{}
	resp.WeiXinPayMsg = l.weixinpayinit
	if l.transantioninfo.CashAccountPayAmount != 0 {
		resp.NeedCashAccountPay = true
	}
	if l.transantioninfo.WexinPayAmount != 0 {
		resp.NeedWeiXinPay = true
	}
	resp.WeiXinPayAmmount = l.transantioninfo.WexinPayAmount
	resp.CashPayAmmount = l.transantioninfo.CashAccountPayAmount
	return resp
}
func (l *PayLogic) transactionend() {
	if l.req.NeedCashAccount {
		wxammount, cashammount, _, needcash, ok := l.CalculatePayAmmount(l.req.Ammount)
		if !ok {
			l.weixinpayall()
		} else {
			if !needcash {
				l.weixinpayall()
			} else {
				l.transantioninfo.NeedCashAccount = 1
				l.transantioninfo.WexinPayAmount = wxammount
				l.transantioninfo.CashAccountPayAmount = cashammount
			}
		}
	} else {
		l.weixinpayall()
	}
}
func (l *PayLogic) weixinpayall() {
	l.weixinpayinit = l.Weixinpayinit(l.transantioninfo.OutTradeNo, l.transantioninfo.Amount)
	l.transantioninfo.WexinPayAmount = l.transantioninfo.Amount
}
func (l *PayLogic) transactioninfoinit() {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "2099-01-01 00:00:00")
	transantioninfo := &cachemodel.TransactionInfo{}
	transantioninfo.OutTradeNo = randStr(32)
	transantioninfo.Phone = l.req.Phone
	transantioninfo.OrderSn = l.req.OrderSn
	transantioninfo.NeedCashAccount = bool2int(l.req.NeedCashAccount)
	transantioninfo.Amount = l.req.Ammount
	transantioninfo.TransactionType = l.req.TransactionType
	transantioninfo.Status = 0
	transantioninfo.CreateOrderTime = time.Now()
	transantioninfo.CashAccountPaymentTime = inittime
	transantioninfo.WexinPaymentTime = inittime
	l.transantioninfo = transantioninfo
}
func int2bool(a int64) bool {
	if a == 0 {
		return false
	} else {
		return true
	}
}
func bool2int(yes bool) int64 {
	if yes {
		return 1
	} else {
		return 0
	}
}
func (l *PayLogic) CalculatePayAmmount(totalammount int64) (wxammount, cashammount int64, needweixin, needcash, ok bool) {
	cash, _ := l.svcCtx.CashAccount.FindOneByPhoneNoCach(l.ctx, l.ctx.Value("phone").(string))
	if cash == nil {
		return totalammount, 0, true, false, true
	} else {
		if cash.Balance >= totalammount {
			return 0, totalammount, false, true, true
		} else {
			return totalammount - cash.Balance, cash.Balance, true, true, true
		}
	}
}
func (l *PayLogic) Weixinpayinit(OutTradeNo string, ammount int64) *types.WeiXinPayMsg {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	// 得到prepay_id，以及调起支付所需的参数和签名
	payment, result, err := jssvc.PrepayWithRequestPayment(l.ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(l.svcCtx.Config.WxConf.AppId),
			Mchid:       core.String(l.svcCtx.Config.WxConf.MchID),
			Description: core.String("沾还是不沾芥末，这是一个问题"),
			OutTradeNo:  core.String(OutTradeNo),
			Attach:      core.String(randStr(16)),
			NotifyUrl:   core.String(l.svcCtx.Config.ServerInfo.Url + "/payrecall/tellmeso"),
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
	return &types.WeiXinPayMsg{PaySign: paySign, NonceStr: nonceStr, TimeStamp: timestampsec, Package: packagestr, SignType: signType}
}
