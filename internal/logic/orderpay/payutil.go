package orderpay

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"time"
)

type PayUtilLogic struct {
	logx.Logger
	ctx          context.Context
	svcCtx       *svc.ServiceContext
	phone        string
	openid       string
	PayInit      *types.PayInit
	PayInfo      *cachemodel.PayInfo
	WeChatPayMsg *types.WeChatPayMsg
	wcu          *WeChatUtilLogic
	OrderList    []*cachemodel.Order
	Wexin        int64
	Cash         int64
}

func NewPayUtilLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayUtilLogic {
	return &PayUtilLogic{
		Logger:       logx.WithContext(ctx),
		ctx:          ctx,
		svcCtx:       svcCtx,
		phone:        ctx.Value("phone").(string),
		openid:       ctx.Value("openid").(string),
		WeChatPayMsg: &types.WeChatPayMsg{},
		wcu:          NewWeChatUtilLogic(ctx, svcCtx),
	}
}
func (l *PayUtilLogic) Payorder(PayInit *types.PayInit, OrderList []*cachemodel.Order) (resp *types.PayMsg, ol []*cachemodel.Order, success bool) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	l.PayInit = PayInit
	l.OrderList = OrderList
	l.payinfoinit()
	l.payinfofinish()
	resp = l.db2resp()
	if !l.OrdersEnd() {
		return nil, nil, false
	}
	sn, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, PayInit.OutTradeSn)
	if sn != nil {
		l.PayInfo.Id = sn.Id
		l.svcCtx.PayInfo.Update(l.ctx, l.PayInfo)
	} else {
		l.svcCtx.PayInfo.Insert(l.ctx, l.PayInfo)
	}
	return resp, OrderList, true

}
func (l *PayUtilLogic) OrdersEnd() bool {
	l.Wexin = l.PayInfo.WexinPayAmount
	l.Cash = l.PayInfo.CashAccountPayAmount
	for i, order := range l.OrderList {
		end, ok := l.OrderEnd(order)
		if ok {
			l.OrderList[i] = end
		} else {
			return false
		}
	}
	return true
}
func (l *PayUtilLogic) OrderEnd(order *cachemodel.Order) (*cachemodel.Order, bool) {
	if order.ActualAmount <= l.Wexin {
		order.WexinPayAmount = order.ActualAmount
		l.Wexin = l.Wexin - order.WexinPayAmount
		return order, true
	} else {
		order.WexinPayAmount = l.Wexin
		l.Wexin = 0
	}
	cashamount := order.ActualAmount - order.WexinPayAmount
	if cashamount > 0 {
		order.CashAccountPayAmount = cashamount
		l.Cash = l.Cash - cashamount
	}
	if l.Cash < 0 {
		return order, false
	}
	return order, true
}

func (l *PayUtilLogic) payinfoinit() {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "2099-01-01 00:00:00")
	PayInfo := &cachemodel.PayInfo{}
	PayInfo.OutTradeNo = l.PayInit.OutTradeSn
	PayInfo.Phone = l.PayInit.Phone
	PayInfo.TotleAmount = l.PayInit.TotleAmmount
	PayInfo.TransactionType = l.PayInit.TransactionType
	PayInfo.Status = 0
	PayInfo.CreateOrderTime = time.Now()
	PayInfo.CashAccountPaymentTime = inittime
	PayInfo.WexinPaymentTime = inittime
	l.PayInfo = PayInfo
}
func (l *PayUtilLogic) payinfofinish() {
	if l.PayInit.NeedCashAccount {
		wxammount, cashammount, _, _, ok := l.CalculatePayAmmount(l.PayInit.TotleAmmount)
		if !ok {
			l.weixinpayall()
		} else {
			l.PayInfo.WexinPayAmount = wxammount
			l.PayInfo.CashAccountPayAmount = cashammount
			if wxammount > 0 {
				l.WeChatPayMsg = l.wcu.Weixinpayinit(l.PayInfo.OutTradeNo, l.PayInfo.WexinPayAmount)
			}
		}
	} else {
		l.weixinpayall()
	}
}

func (l *PayUtilLogic) CalculatePayAmmount(totalammount int64) (wxammount, cashammount int64, needweixin, needcash, ok bool) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	cash, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, l.phone)
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
func (l *PayUtilLogic) weixinpayall() {
	l.WeChatPayMsg = l.wcu.Weixinpayinit(l.PayInfo.OutTradeNo, l.PayInfo.TotleAmount)
	l.PayInfo.WexinPayAmount = l.PayInfo.TotleAmount
}
func (l *PayUtilLogic) db2resp() *types.PayMsg {
	resp := &types.PayMsg{}
	resp.WeChatPayMsg = l.WeChatPayMsg
	if l.PayInfo.CashAccountPayAmount != 0 {
		resp.NeedCashAccountPay = true
	}
	if l.PayInfo.WexinPayAmount != 0 {
		resp.NeedWeChatPay = true
	}
	resp.WeChatPayAmmount = l.PayInfo.WexinPayAmount
	resp.CashPayAmmount = l.PayInfo.CashAccountPayAmount
	return resp
}
