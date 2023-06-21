package shoppingcart

import (
	"context"
	"encoding/json"
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
	tinyproductlist := make([]*types.ProductTiny, 0)
	if scinfo != nil {
		json.Unmarshal([]byte(scinfo.ShoppingCart), &tinyproductlist)
	}
	return &types.GetShoppingCartResp{Code: "10000", Msg: "success", Data: &types.ShoppingCart{GoodsList: tinyproductlist}}, nil

}
