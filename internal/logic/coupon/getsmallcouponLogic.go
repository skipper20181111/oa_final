package coupon

import (
	"context"
	"encoding/json"
	"math/rand"
	"oa_final/cachemodel"
	"strconv"
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
	infolist := make([]*types.CouponInfo, 0)
	if couponbyphone == nil {
		couponmap = getnewcoupon(couponmap, couponinfomap)
		couponmapstr, _ := json.Marshal(couponmap)
		infolist, couponmap = getinfolist(couponmap, couponinfomap)
		l.svcCtx.UserCoupon.Insert(l.ctx, &cachemodel.UserCoupon{Phone: l.ctx.Value("phone").(string), CouponIdMap: string(couponmapstr)})
		return &types.GetSmallCouponResp{Code: "10000", Msg: "列表如下", Data: &types.GetSmallCouponRp{CouponInfoList: infolist}}, nil
	} else {
		json.Unmarshal([]byte(couponbyphone.CouponIdMap), &couponmap)
		infolist, couponmap = getinfolist(couponmap, couponinfomap)
		couponmapstr, _ := json.Marshal(couponmap)
		couponbyphone.CouponIdMap = string(couponmapstr)
		l.svcCtx.UserCoupon.Update(l.ctx, couponbyphone)
		return &types.GetSmallCouponResp{Code: "10000", Msg: "列表如下", Data: &types.GetSmallCouponRp{CouponInfoList: infolist}}, nil
	}
	return &types.GetSmallCouponResp{Code: "4004", Msg: "无"}, nil
}
func getnewcoupon(couponmap map[int64]map[string]*types.CouponStoreInfo, couponinfomap map[int64]*types.CouponInfo) map[int64]map[string]*types.CouponStoreInfo {
	newusercouponid := []int64{10000, 10001, 10002, 10003, 10004, 10005, 10006, 10007, 10008, 10009, 10010}
	for _, cid := range newusercouponid {
		info, ok := couponinfomap[cid]
		if ok {
			couponmap[cid] = make(map[string]*types.CouponStoreInfo)
			couponmap[cid][strconv.FormatInt(time.Now().UnixNano()+rand.Int63n(10000), 10)] = &types.CouponStoreInfo{CouponId: cid, DisabledTime: time.Now().Add(time.Hour * time.Duration(24*info.EfficientPeriod)).Format("2006-01-02 15:04:05")}
		} else {
			return couponmap
		}
	}
	return couponmap
}
func getinfolist(couponmap map[int64]map[string]*types.CouponStoreInfo, couponinfomap map[int64]*types.CouponInfo) ([]*types.CouponInfo, map[int64]map[string]*types.CouponStoreInfo) {
	infolist := make([]*types.CouponInfo, 0)
	for cid, uuidmap := range couponmap {
		if _, ok := couponinfomap[cid]; !ok {
			delete(couponmap, cid)
			continue
		}
		if len(uuidmap) == 0 {
			delete(couponmap, cid)
			continue
		}
		chilemap := &uuidmap
		for uuid, storeInfo := range uuidmap {
			disabletime, _ := time.Parse("2006-01-02 15:04:05", storeInfo.DisabledTime)
			if disabletime.Before(time.Now()) {
				delete(*chilemap, uuid)
			} else {
				couponinfo := *(couponinfomap[cid])
				couponinfo.CouponUUID = uuid
				couponinfo.DisabledTime = storeInfo.DisabledTime
				infolist = append(infolist, &couponinfo)
			}
		}
		couponmap[cid] = *chilemap
	}
	for i := 0; i < len(infolist); i++ {
		for j := 1; j < len(infolist)-i; j++ {
			if timesmaller(infolist[j].DisabledTime, infolist[j-1].DisabledTime) {
				//交换
				infolist[j], infolist[j-1] = infolist[j-1], infolist[j]
			}
		}
	}
	return infolist, couponmap
}
func timesmaller(a, b string) bool {
	atime, _ := time.Parse("2006-01-02 15:04:05", a)
	btime, _ := time.Parse("2006-01-02 15:04:05", b)
	return atime.Before(btime)
}
