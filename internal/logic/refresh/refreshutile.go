package refresh

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"strings"
	"time"
)

type RefreshUtilLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshUtilLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshUtilLogic {
	return &RefreshUtilLogic{
		Logger: logx.WithContext(ctx),
		ctx:    context.Background(),
		svcCtx: svcCtx,
	}
}
func (l *RefreshUtilLogic) MissionList() {
	MissionList := make([]*types.Mission, 0)
	missions, _ := l.svcCtx.Mission.FindAll(l.ctx)
	if missions != nil && len(missions) >= 1 {
		for _, mission := range missions {
			MissionList = append(MissionList, missiondb2info(mission))
		}
		l.svcCtx.LocalCache.Set(svc.MissionListKey, MissionList)
	}
}
func missiondb2info(db *cachemodel.Mission) *types.Mission {
	minfo := &types.Mission{}
	minfo.MissionId = db.MissionId
	minfo.Count = db.ConsumeCount
	describelist := strings.Split(db.Describe, "#")
	minfo.Describe = describelist
	return minfo
}
func (l RefreshUtilLogic) SfPrice() {
	sfPrices, _ := l.svcCtx.SfPrice.FindAll(l.ctx)
	if len(sfPrices) > 30 {
		SfPriceMap := make(map[string]*types.SfPriceInfo)
		for _, sfPrice := range sfPrices {
			SfPriceMap[sfPrice.Province] = &types.SfPriceInfo{
				Province: sfPrice.Province,
				FirstKG:  sfPrice.StartKg,
				PerKG:    sfPrice.PerKg,
			}
		}
		l.svcCtx.LocalCache.Set(svc.SfPrice, SfPriceMap)
	}
}
func (l RefreshUtilLogic) Products() {
	AllProductsList, _ := l.svcCtx.Product.FindAll(l.ctx)
	if AllProductsList != nil && len(AllProductsList) > 0 {
		ProductDBMap := make(map[int64]*cachemodel.Product)
		QuantityInfoDBList := make(map[int64][]*types.QuantityInfoDB)
		ProductQuantityInfoDB := make(map[int64]map[string]*types.QuantityInfoDB)
		for _, product := range AllProductsList {
			qinfoList := &types.QuantityInfoDBList{}
			json.Unmarshal([]byte(product.QuantityPriceCutInfo), &qinfoList)
			QuantityInfoDBList[product.Pid] = qinfoList.InfoList
			ProductDBMap[product.Pid] = product
			qinfo := make(map[string]*types.QuantityInfoDB)
			for _, db := range qinfoList.InfoList {
				qinfo[db.Name] = db
			}
			ProductQuantityInfoDB[product.Pid] = qinfo
		}

		l.svcCtx.LocalCache.Set(svc.ProductsMap, ProductDBMap)
		l.svcCtx.LocalCache.Set(svc.ProductQuantityInfoDB, ProductQuantityInfoDB)
		l.svcCtx.LocalCache.Set(svc.QuantityInfoDBList, QuantityInfoDBList)
	}
}

func (l *RefreshUtilLogic) RechargeProduct() bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	all, _ := l.svcCtx.RechargeProduct.FindAll(l.ctx)
	rcpmap := make(map[int64]*cachemodel.RechargeProduct)
	if all != nil && len(all) >= 1 {
		for _, product := range all {
			rcpmap[product.Rpid] = product
		}
		l.svcCtx.LocalCache.Set(svc.RechargeProductKey, rcpmap)
	}

	return true
}
func (l *RefreshUtilLogic) StarMall() bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	// 开始starmall的缓存
	findAll, _ := l.svcCtx.StarMallLongList.FindAll(l.ctx)
	StarMallMap := make(map[int64]*cachemodel.StarmallLonglist, 0)
	if findAll != nil && len(findAll) >= 1 {
		for _, longlist := range findAll {
			StarMallMap[longlist.ProductId] = longlist
		}
		l.svcCtx.LocalCache.Set(svc.StarMallMap, StarMallMap)
	}

	return true
}
func (l *RefreshUtilLogic) Coupon() bool {
	coupons, _ := l.svcCtx.Coupon.FindAll(l.ctx)
	couponmap := make(map[int64]*cachemodel.Coupon)
	cinfomap := make(map[int64]*types.CouponInfo)
	if coupons != nil {
		for _, coupon := range coupons {
			if coupon.Discount == 0 && coupon.MinPoint == 0 && coupon.Cut == 0 {
				continue
			}
			couponmap[coupon.CouponId] = coupon
			info := coupondb2info(coupon)
			cinfomap[info.CouponId] = info
		}
		l.svcCtx.LocalCache.Set(svc.CouponMapKey, couponmap)
		l.svcCtx.LocalCache.Set(svc.CouponInfoMapKey, cinfomap)
	}

	return true
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
func coupondb2info(coupon *cachemodel.Coupon) *types.CouponInfo {
	cinfo := &types.CouponInfo{}
	cinfo.CouponId = coupon.CouponId
	cinfo.Type = coupon.TypeZh
	cinfo.Title = coupon.Name
	cinfo.LeastConsume = float64(coupon.MinPoint) / 100
	cinfo.AvailableRange = "全场通用"
	rules := make([]string, 0)
	json.Unmarshal([]byte(coupon.Note), &rules)
	cinfo.Rules = rules
	cinfo.Cut = float64(coupon.Cut) / 100
	cinfo.Discount = coupon.Discount
	cinfo.TypeCode = coupon.Type
	cinfo.EfficientPeriod = coupon.EfficientPeriod

	exchangenote := make([]string, 0)
	json.Unmarshal([]byte(coupon.ExchangeNotes), &exchangenote)
	cinfo.ExchangeNotes = exchangenote
	return cinfo
}
