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
	productsInfoMap := make(map[int64]*types.ProductInfo)
	productsMap := make(map[int64]*cachemodel.Product)
	for _, product := range productList {
		info := product2info(product)
		productsMap[product.Pid] = product
		productsInfoMap[product.Pid] = &info
	}
	l.svcCtx.LocalCache.Set(svc.ProductsInfoMap, productsInfoMap)
	l.svcCtx.LocalCache.Set(svc.ProductsMap, productsMap)

	all, _ := l.svcCtx.RechargeProduct.FindAll(l.ctx)
	rcpmap := make(map[int64]*cachemodel.RechargeProduct)
	if all != nil && len(all) >= 1 {
		for _, product := range all {
			rcpmap[product.Rpid] = product
		}
	}
	l.svcCtx.LocalCache.Set(svc.RechargeProductKey, rcpmap)

	// 开始starmall的缓存
	findAll, _ := l.svcCtx.StarMallLongList.FindAll(l.ctx)
	StarMallMap := make(map[int64]*cachemodel.StarmallLonglist, 0)
	if findAll != nil && len(findAll) >= 1 {
		for _, longlist := range findAll {
			StarMallMap[longlist.ProductId] = longlist
		}
	}
	l.svcCtx.LocalCache.Set(svc.StarMallMap, StarMallMap)

	return &types.RefreshResp{Code: "10000", Msg: "刷新成功"}, err
}
func product2info(product *cachemodel.Product) (info types.ProductInfo) {
	info.ProductId = product.Pid
	info.Product_title = product.ProductTitle
	info.Picture = product.Picture
	info.Status = int(product.Status)
	info.Reserve_time = time.Time.String(product.ReserveTime)
	info.Sale = int(product.Sale)
	info.Promotion_price = float64(product.PromotionPrice) / 100
	info.Original_price = float64(product.OriginalPrice) / 100
	info.Cut_price = float64(product.CutPrice) / 100
	info.Description = product.Description
	info.Unit = product.Unit
	info.Weight = product.Weight
	info.Attribute = product.Attribute
	return info

}
