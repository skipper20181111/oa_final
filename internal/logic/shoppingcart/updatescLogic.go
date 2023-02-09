package shoppingcart

import (
	"context"
	"encoding/json"
	"oa_final/cachemodel"
	"oa_final/internal/logic/refresh"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatescLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatescLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatescLogic {
	return &UpdatescLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatescLogic) Updatesc(req *types.UpdateShoppingCartRes) (resp *types.UpdateShoppingCartResp, err error) {
	userphone := l.ctx.Value("phone").(string)
	shoppingCart, err := json.Marshal(req.ShopCartIdList)
	scphone, err := l.svcCtx.UserShopping.FindOneByPhone(l.ctx, userphone)
	if scphone == nil && err.Error() == "notfind" {
		l.svcCtx.UserShopping.Insert(l.ctx, &cachemodel.UserShoppingCart{Phone: userphone, ShoppingCart: string(shoppingCart)})
	} else if scphone != nil {
		l.svcCtx.UserShopping.UpdateByPhone(l.ctx, userphone, string(shoppingCart))
	} else {
		goodsList := make([]*types.ProductInfo, 0)
		return &types.UpdateShoppingCartResp{Code: "10000", Msg: "success", Data: &types.ShoppingCart{GoodsList: goodsList}}, nil
	}

	scinfo, err := l.svcCtx.UserShopping.FindOneByPhone(l.ctx, userphone)
	if scinfo == nil {
		goodsList := make([]*types.ProductInfo, 0)
		return &types.UpdateShoppingCartResp{Code: "10000", Msg: "success", Data: &types.ShoppingCart{GoodsList: goodsList}}, nil
	}
	tinyproductlist := make([]types.ProductTiny, 0)
	json.Unmarshal([]byte(scinfo.ShoppingCart), &tinyproductlist)
	//for i, productTiny := range tinyproductlist {
	//	productTiny.
	//}
	PMcache, ok := l.svcCtx.LocalCache.Get(refresh.ProductsMap)
	if !ok {
		return &types.UpdateShoppingCartResp{Code: "4004", Msg: "此地无缓存"}, nil
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
	return &types.UpdateShoppingCartResp{Code: "10000", Msg: "success", Data: &types.ShoppingCart{GoodsList: goodsList}}, nil
}
