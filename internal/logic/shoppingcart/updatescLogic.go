package shoppingcart

import (
	"context"
	"encoding/json"
	"oa_final/cachemodel"
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
	if scphone == nil {
		l.svcCtx.UserShopping.Insert(l.ctx, &cachemodel.UserShoppingCart{Phone: userphone, ShoppingCart: string(shoppingCart)})
	} else {
		l.svcCtx.UserShopping.UpdateByPhone(l.ctx, userphone, string(shoppingCart))
	}
	scinfo, err := l.svcCtx.UserShopping.FindOneByPhone(l.ctx, userphone)
	tinyproductlist := make([]*types.ProductTiny, 0)
	if scinfo != nil {
		json.Unmarshal([]byte(scinfo.ShoppingCart), &tinyproductlist)
	}

	return &types.UpdateShoppingCartResp{Code: "10000", Msg: "success", Data: &types.ShoppingCart{GoodsList: tinyproductlist}}, nil
}
