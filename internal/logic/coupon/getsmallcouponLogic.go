package coupon

import (
	"context"
	"encoding/json"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetsmallcouponLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	phone  string
}

func NewGetsmallcouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetsmallcouponLogic {
	return &GetsmallcouponLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		phone:  ctx.Value("phone").(string),
	}
}

func (l *GetsmallcouponLogic) Getsmallcoupon(req *types.GetSmallCouponRes) (resp *types.GetSmallCouponResp, err error) {
	get, ok := l.svcCtx.LocalCache.Get(svc.CouponInfoMapKey)
	if !ok {
		return &types.GetSmallCouponResp{Code: "4004", Msg: "无缓存"}, nil
	}
	couponinfomap := get.(map[int64]*types.CouponInfo)
	couponbyphone, err := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, l.phone)
	couponmap := make(map[int64]map[string]*types.CouponStoreInfo)
	if couponbyphone == nil {
		couponmapstr, _ := json.Marshal(couponmap)
		l.svcCtx.UserCoupon.Insert(l.ctx, &cachemodel.UserCoupon{Phone: l.ctx.Value("phone").(string), CouponIdMap: string(couponmapstr)})
		return &types.GetSmallCouponResp{Code: "10000", Msg: "列表如下", Data: &types.GetSmallCouponRp{CouponInfoList: make([]*types.CouponInfo, 0)}}, nil
	} else {
		json.Unmarshal([]byte(couponbyphone.CouponIdMap), &couponmap)
		infolist := make([]*types.CouponInfo, 0)
		for cid, uuidmap := range couponmap {
			if _, ok := couponinfomap[cid]; !ok {
				delete(couponmap, cid)
				continue
			}
			chilemap := &uuidmap
			for uuid, storeInfo := range uuidmap {
				disabletime, _ := time.Parse("2006-01-02 15:04:05", storeInfo.DisabledTime)
				if disabletime.Before(time.Now()) {
					delete(*chilemap, uuid)
				} else {
					couponinfomap[cid].CouponUUID = uuid
					couponinfomap[cid].DisabledTime = storeInfo.DisabledTime
					infolist = append(infolist, couponinfomap[cid])
				}
			}
			couponmap[cid] = *chilemap
		}
		couponmapstr, _ := json.Marshal(couponmap)
		couponbyphone.CouponIdMap = string(couponmapstr)
		l.svcCtx.UserCoupon.Update(l.ctx, couponbyphone)
		return &types.GetSmallCouponResp{Code: "10000", Msg: "列表如下", Data: &types.GetSmallCouponRp{CouponInfoList: infolist}}, nil
	}
	return &types.GetSmallCouponResp{Code: "4004", Msg: "无"}, nil
}
