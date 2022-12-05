package refresh

import (
	"context"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshPLLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

const ProductsMap = "ProductsMap"

func NewRefreshPLLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshPLLogic {
	return &RefreshPLLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshPLLogic) RefreshPL() (resp *types.RefreshResp, err error) {
	productList, err := l.svcCtx.Product.FindAll(l.ctx)
	if err != nil {
		return &types.RefreshResp{Code: "4004", Msg: "失败"}, err
	}
	productsMap := make(map[int64]*types.ProductInfo)
	for _, product := range productList {
		info := product2info(product)
		productsMap[product.Pid] = &info
	}
	l.svcCtx.LocalCache.Set(ProductsMap, productsMap)
	return &types.RefreshResp{Code: "10000", Msg: "刷新成功"}, err
}
func product2info(product *cachemodel.Product) (info types.ProductInfo) {
	info.ProductId = product.Pid
	info.Product_title = product.ProductTitle
	info.Picture = product.Picture
	info.Status = int(product.Status)
	info.Reserve_time = time.Time.String(product.ReserveTime)
	info.Sale = int(product.Sale)
	info.Promotion_price = product.PromotionPrice
	info.Original_price = product.OriginalPrice
	info.Cut_price = product.CutPrice
	info.Description = product.Description
	info.Unit = product.Unit
	info.Weight = product.Weight
	info.Attribute = product.Attribute
	return info

}
