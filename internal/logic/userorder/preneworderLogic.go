package userorder

import (
	"context"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreneworderLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
}

func NewPreneworderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreneworderLogic {
	return &PreneworderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreneworderLogic) Preneworder(req *types.NewOrderRes) (resp *types.PreNewOrderResp, err error) {
	l.userphone = l.ctx.Value("phone").(string)
	if len(req.ProductTinyList) == 0 {
		return &types.PreNewOrderResp{Code: "10000", Msg: "商品列表为空", Data: &types.PreNewOrderRp{PreOrderInfo: nil}}, nil
	}
	PMcache, ok := l.svcCtx.LocalCache.Get(svc.ProductsMap)
	if !ok {
		return &types.PreNewOrderResp{Code: "4004", Msg: "服务器查找商品列表失败"}, nil
	}
	productsMap := PMcache.(map[int64]*cachemodel.Product)
	lu := NewLogic(l.ctx, l.svcCtx)

	orderinfo := OrderDb2Preinfo(lu.Order2db(req, productsMap, UseCache(true)))
	return &types.PreNewOrderResp{Code: "10000", Msg: "结算完成，请下订单", Data: &types.PreNewOrderRp{PreOrderInfo: orderinfo}}, nil
}
