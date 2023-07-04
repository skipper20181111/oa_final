package orderpay

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
	oul       *OrderUtilLogic
}

func NewOrderoptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderoptionLogic {
	return &OrderoptionLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		userphone: ctx.Value("phone").(string),
		oul:       NewOrderUtilLogic(ctx, svcCtx),
	}
}

func (l *OrderoptionLogic) Orderoption(req *types.OrderOptionRes) (resp *types.OrderOptionResp, err error) {
	balance := float64(0)
	OriginalAmount, PromotionAmount := l.oul.GetAmount(req.ProductTinyList)
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
	couponstoremap := make(map[int64]map[string]*types.CouponStoreInfo)
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
