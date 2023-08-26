package coupon

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"oa_final/cachemodel"
	"oa_final/internal/logic/orderpay"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type StarmallcouponorderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	phone  string
}

func NewStarmallcouponorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StarmallcouponorderLogic {
	return &StarmallcouponorderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		phone:  ctx.Value("phone").(string),
	}
}

func (l *StarmallcouponorderLogic) Starmallcouponorder(req *types.StarMallCouponOrderRes) (resp *types.StarMallCouponOrderResp, err error) {
	get, ok := l.svcCtx.LocalCache.Get(svc.CouponInfoMapKey)
	if !ok {
		return &types.StarMallCouponOrderResp{Code: "4004", Msg: "无缓存"}, nil
	}
	couponinfomap := get.(map[int64]*types.CouponInfo)
	couponadd, _ := l.svcCtx.Coupon.FindOneByCouponId(l.ctx, req.Cid)
	userpoints, _ := l.svcCtx.UserPoints.FindOneByPhone(l.ctx, l.phone)
	if userpoints == nil || couponadd == nil || userpoints.AvailablePoints < couponadd.UsePoints || couponadd.UsePoints == 0 {
		return &types.StarMallCouponOrderResp{Code: "10000", Msg: "积分不够"}, nil
	}
	couponbyphone, err := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, l.ctx.Value("phone").(string))
	couponmap := make(map[int64]map[string]*types.CouponStoreInfo)
	if couponbyphone == nil {
		couponmap[req.Cid] = make(map[string]*types.CouponStoreInfo)
		couponmap[req.Cid][strconv.FormatInt(time.Now().UnixNano(), 10)] = &types.CouponStoreInfo{CouponId: req.Cid, DisabledTime: time.Now().Add(time.Hour * time.Duration(24*couponadd.EfficientPeriod)).Format("2006-01-02 15:04:05")}
		couponmapstr, _ := json.Marshal(couponmap)
		l.svcCtx.UserCoupon.Insert(l.ctx, &cachemodel.UserCoupon{
			Phone:       l.ctx.Value("phone").(string),
			CouponIdMap: string(couponmapstr),
		})
		userpoints.AvailablePoints = userpoints.AvailablePoints - couponadd.UsePoints
		l.svcCtx.UserPoints.Update(l.ctx, userpoints)
		l.svcCtx.PointLog.Insert(l.ctx, &cachemodel.PointLog{Date: time.Now(),
			OrderType:     "兑换优惠券",
			OrderSn:       orderpay.GetSha256(fmt.Sprintf("%d%d%s", time.Now().UnixNano(), userpoints.AvailablePoints, couponmapstr)),
			OrderDescribe: "臻星商城兑换优惠券",
			Behavior:      "兑换",
			Phone:         l.phone,
			Balance:       userpoints.AvailablePoints,
			ChangeAmount:  couponadd.UsePoints,
		})
		l.GetOrder(couponadd)
		return &types.StarMallCouponOrderResp{Code: "10000", Msg: "列表如下", Data: &types.GetSmallCouponRp{CouponInfoList: make([]*types.CouponInfo, 0)}}, nil
	} else if couponbyphone != nil {
		json.Unmarshal([]byte(couponbyphone.CouponIdMap), &couponmap)
		_, ok := couponmap[req.Cid]
		if ok {
			couponmap[req.Cid][strconv.FormatInt(time.Now().UnixNano()+rand.Int63n(10000), 10)] = &types.CouponStoreInfo{CouponId: req.Cid, DisabledTime: time.Now().Add(time.Hour * time.Duration(24*couponadd.EfficientPeriod)).Format("2006-01-02 15:04:05")}
		} else {
			couponmap[req.Cid] = make(map[string]*types.CouponStoreInfo)
			couponmap[req.Cid][strconv.FormatInt(time.Now().UnixNano()+rand.Int63n(10000), 10)] = &types.CouponStoreInfo{CouponId: req.Cid, DisabledTime: time.Now().Add(time.Hour * time.Duration(24*couponadd.EfficientPeriod)).Format("2006-01-02 15:04:05")}
		}
		couponmapstr, _ := json.Marshal(couponmap)
		couponbyphone.CouponIdMap = string(couponmapstr)
		l.svcCtx.UserCoupon.Update(l.ctx, couponbyphone)
		userpoints.AvailablePoints = userpoints.AvailablePoints - couponadd.UsePoints
		l.svcCtx.UserPoints.Update(l.ctx, userpoints)
		l.svcCtx.PointLog.Insert(l.ctx, &cachemodel.PointLog{Date: time.Now(),
			OrderType:     "兑换优惠券",
			OrderSn:       orderpay.GetSha256(fmt.Sprintf("%d%d%s", time.Now().UnixNano(), userpoints.AvailablePoints, couponmapstr)),
			OrderDescribe: "臻星商城兑换优惠券",
			Behavior:      "兑换",
			Phone:         l.phone,
			Balance:       userpoints.AvailablePoints,
			ChangeAmount:  couponadd.UsePoints,
		})
		l.GetOrder(couponadd)
		infolist := make([]*types.CouponInfo, 0)
		for cid, uuidmap := range couponmap {
			for uuid, storeInfo := range uuidmap {
				couponinfo := *(couponinfomap[cid])
				couponinfo.CouponUUID = uuid
				couponinfo.DisabledTime = storeInfo.DisabledTime
				infolist = append(infolist, &couponinfo)
			}
		}
		return &types.StarMallCouponOrderResp{Code: "10000", Msg: "列表如下", Data: &types.GetSmallCouponRp{CouponInfoList: infolist}}, nil
	}
	return &types.StarMallCouponOrderResp{Code: "10000", Msg: "无缓存"}, nil
}
func (l StarmallcouponorderLogic) GetOrder(couponinfo *cachemodel.Coupon) {
	order := &cachemodel.Order{}
	order.OrderType = 0
	order.PointsOrder = 1
	order.PointsAmount = couponinfo.UsePoints
	order.OrderStatus = 4
	order.Phone = l.phone
	order.OutTradeNo = orderpay.RandStr(64)
	order.OutRefundNo = orderpay.RandStr(64)
	order.CreateOrderTime = time.Now()
	order.ModifyTime = order.CreateOrderTime
	order.PaymentTime = time.Now()
	order.DeliveryTime = time.Now()
	order.ReceiveTime = time.Now()
	order.CloseTime = time.Now()
	order.OrderSn = orderpay.Getsha512(order.Phone + order.CreateOrderTime.String() + couponinfo.Name + orderpay.RandStr(64))
	order.LogId = time.Now().UnixNano()
	order.ProductInfo = fmt.Sprintf("%s %s * %d %s", couponinfo.TypeZh, couponinfo.Name, 1, "\n")
	l.svcCtx.Order.Insert(l.ctx, order)
}
