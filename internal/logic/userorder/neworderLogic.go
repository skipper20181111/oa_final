package userorder

import (
	"context"
	"oa_final/cachemodel"
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
	useropenid  string
}

func NewNeworderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NeworderLogic {
	return &NeworderLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
	}
}

func (l *NeworderLogic) Neworder(req *types.NewOrderRes) (resp *types.NewOrderResp, err error) {
	UseAccount := false
	if len(req.ProductTinyList) == 0 {
		return &types.NewOrderResp{Code: "4004", Msg: "无商品，订单金额为0", Data: &types.NewOrderRp{}}, nil
	}
	PMcache, ok := l.svcCtx.LocalCache.Get(svc.ProductsMap)
	if !ok {
		return &types.NewOrderResp{Code: "4004", Msg: "服务器查找商品列表失败"}, nil
	}
	productsMap := PMcache.(map[int64]*cachemodel.Product)
	lu := NewLogic(l.ctx, l.svcCtx)
	order := lu.Order2db(req, productsMap, UseCache(false))
	l.svcCtx.UserOrder.Insert(l.ctx, order)
	l.userorder = order
	payl := NewPayLogic(l.ctx, l.svcCtx)
	payorder, success := payl.Payorder(&types.TransactionInit{TransactionType: "普通商品", OrderSn: l.userorder.OrderSn, OutTradeSn: l.userorder.OutTradeNo, NeedCashAccount: req.UseCashFirst, Ammount: l.userorder.ActualAmount, Phone: l.userphone})
	if !success {
		return &types.NewOrderResp{Code: "4004", Msg: "fatal error"}, nil
	}
	if l.usepoint || l.usecoupon || payorder.NeedCashAccountPay {
		UseAccount = true
	}
	order.WexinPayAmount = payorder.WeiXinPayAmmount
	order.CashAccountPayAmount = payorder.CashPayAmmount
	l.svcCtx.UserOrder.UpdateByOrderSn(l.ctx, order)
	neworderrp := types.NewOrderRp{OrderInfo: OrderDb2info(order, nil), UseAccount: UseAccount, UseWechatPay: payorder.NeedWeiXinPay, WeiXinPayMsg: payorder.WeiXinPayMsg}
	return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &neworderrp}, nil
}
