package userorder

import (
	"context"
	"fmt"
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
}

func NewNeworderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NeworderLogic {
	return &NeworderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
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
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, order.OrderSn)
	if sn2order == nil {
		fmt.Println(err.Error())
		return &types.NewOrderResp{Code: "4004", Msg: "数据库失效"}, nil
	}
	l.userorder = sn2order

	payl := NewPayLogic(l.ctx, l.svcCtx)
	payorder, success := payl.Payorder(&types.TransactionInit{TransactionType: "普通商品", OrderSn: l.userorder.OrderSn, NeedCashAccount: req.UseCashFirst, Ammount: l.userorder.ActualAmount, Phone: l.userphone})
	if success {
		if l.usepoint || l.usecoupon || payorder.NeedCashAccountPay {
			UseAccount = true
		}
		info, _ := l.svcCtx.TransactionInfo.FindOneByOrderSn(l.ctx, sn2order.OrderSn)
		neworderrp := types.NewOrderRp{OrderInfo: OrderDb2info(sn2order, info), UseAccount: UseAccount, UseWechatPay: true, WeiXinPayMsg: payorder.WeiXinPayMsg}
		return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &neworderrp}, nil
	}
	return &types.NewOrderResp{Code: "4004", Msg: "支付失败"}, nil
}
