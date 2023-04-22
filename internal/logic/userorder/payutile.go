package userorder

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
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
	orderutile      *Logic
	WeChatUtilLogic *WeChatUtilLogic
}

func NewPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayLogic {
	return &PayLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		orderutile:      NewLogic(ctx, svcCtx),
		WeChatUtilLogic: NewWeChatUtilLogic(ctx, svcCtx),
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
		l.transantioninfo.Id = sn.Id
		l.svcCtx.TransactionInfo.Update(l.ctx, l.transantioninfo)
	} else {
		l.svcCtx.TransactionInfo.Insert(l.ctx, l.transantioninfo)
	}
	return resp, true

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
				l.weixinpayinit = l.WeChatUtilLogic.Weixinpayinit(l.transantioninfo.OutTradeNo, l.transantioninfo.WexinPayAmount)
			}
		}
	} else {
		l.weixinpayall()
	}
}
func (l *PayLogic) weixinpayall() {
	l.weixinpayinit = l.WeChatUtilLogic.Weixinpayinit(l.transantioninfo.OutTradeNo, l.transantioninfo.Amount)
	l.transantioninfo.WexinPayAmount = l.transantioninfo.Amount
	l.transantioninfo.NeedCashAccount = 0
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
