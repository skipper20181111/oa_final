package shoppingcart

import (
	"context"
	"encoding/json"
	"oa_final/internal/logic/refresh"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetscLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetscLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetscLogic {
	return &GetscLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetscLogic) Getsc(req *types.GetShoppingCartRes) (resp *types.GetShoppingCartResp, err error) {
	userphone := l.ctx.Value("phone").(string)
	scinfo, err := l.svcCtx.UserShopping.FindOneByPhone(l.ctx, userphone)
	if scinfo == nil {
		return &types.GetShoppingCartResp{Code: "10000", Msg: "success", Data: &types.ShoppingCart{GoodsList: make([]*types.ProductInfo, 0)}}, nil
	}
	tinyproductlist := make([]types.ProductTiny, 0)
	json.Unmarshal([]byte(scinfo.ShoppingCart), &tinyproductlist)
	PMcache, ok := l.svcCtx.LocalCache.Get(refresh.ProductsMap)
	if !ok {
		return &types.GetShoppingCartResp{Code: "4004", Msg: "此地无缓存"}, nil
	}
	productsMap := PMcache.(map[int64]*types.ProductInfo)
	goodsList := make([]*types.ProductInfo, 0)
	for _, tiny := range tinyproductlist {
		info, ok := productsMap[tiny.PId]
		if !ok {
			continue
		}
		info.Amount = tiny.Amount
		goodsList = append(goodsList, productsMap[tiny.PId])
	}
	return &types.GetShoppingCartResp{Code: "10000", Msg: "success", Data: &types.ShoppingCart{GoodsList: goodsList}}, nil

}
