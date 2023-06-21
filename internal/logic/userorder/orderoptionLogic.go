package userorder

import (
	"context"
	"encoding/json"
	"oa_final/cachemodel"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderoptionLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
}

func NewOrderoptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderoptionLogic {
	return &OrderoptionLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		userphone: ctx.Value("phone").(string),
	}
}

func (l *OrderoptionLogic) Orderoption(req *types.OrderOptionRes) (resp *types.OrderOptionResp, err error) {
	PMcache, ok := l.svcCtx.LocalCache.Get(svc.ProductsMap)
	if !ok {
		return &types.OrderOptionResp{Code: "4004", Msg: "服务器查找商品列表失败"}, nil
	}
	productsMap := PMcache.(map[int64]*cachemodel.Product)
	OriginalAmount := int64(0)
	PromotionAmount := int64(0)
	balance := float64(0)
	couponstoremap := make(map[int64]map[string]*types.CouponStoreInfo)
	for _, tiny := range req.ProductTinyList {
		PromotionAmount = PromotionAmount + productsMap[tiny.PId].PromotionPrice*int64(tiny.Amount)
		OriginalAmount = OriginalAmount + productsMap[tiny.PId].OriginalPrice*int64(tiny.Amount)
	}
	cach, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, l.userphone)
	if cach != nil {
		balance = float64(cach.Balance) / 100
	}
	get, ok := l.svcCtx.LocalCache.Get(svc.CouponInfoMapKey)
	get2, ok2 := l.svcCtx.LocalCache.Get(svc.CouponMapKey)
	if !ok || !ok2 {
		return &types.OrderOptionResp{Code: "4004", Msg: "缓存失效"}, nil
	}
	couponmap := get2.(map[int64]*cachemodel.Coupon)
	couponinfomap := get.(map[int64]*types.CouponInfo)
	infolist := make([]*types.CouponInfo, 0)
	userCoupon, _ := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, l.userphone)
	if userCoupon != nil {
		json.Unmarshal([]byte(userCoupon.CouponIdMap), &couponstoremap)
		for cid, uuidmap := range couponstoremap {
			for uuid, storeInfo := range uuidmap {
				if ifcouponuseable(couponmap[cid], storeInfo.DisabledTime, PromotionAmount) {
					couponinfomap[cid].CouponUUID = uuid
					couponinfomap[cid].DisabledTime = storeInfo.DisabledTime
					infolist = append(infolist, couponinfomap[cid])
				}
			}
		}
	}
	originalamount := float64(OriginalAmount) / 100
	promotionamount := float64(PromotionAmount) / 100
	return &types.OrderOptionResp{Code: "10000", Msg: "success", Data: &types.OrderOptionRp{PromotionAmount: promotionamount, OriginalAmount: originalamount, AvailableBalance: balance, AvailableCoupon: infolist}}, nil
}
