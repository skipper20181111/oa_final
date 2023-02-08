package userorder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"oa_final/internal/logic/refresh"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreneworderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreneworderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreneworderLogic {
	return &PreneworderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreneworderLogic) Preneworder(req *types.PreNewOrderRes) (resp *types.PreNewOrderResp, err error) {
	if l.ctx.Value("openid") != req.OpenId || l.ctx.Value("phone") != req.Phone {
		return &types.PreNewOrderResp{
			Code: "4004",
			Msg:  "请勿使用其他用户的token",
		}, nil
	}
	if len(req.ProductTinyList) == 0 {
		return &types.PreNewOrderResp{Code: "10000", Msg: "商品列表为空", Data: &types.PreNewOrderRp{PreOrderInfo: nil}}, nil
	}
	PMcache, ok := l.svcCtx.LocalCache.Get(refresh.ProductsMap)
	if !ok {
		return &types.PreNewOrderResp{Code: "4004", Msg: "服务器查找商品列表失败"}, nil
	}
	productsMap := PMcache.(map[int64]*types.ProductInfo)
	orderinfo := l.order2orderInfo(req, productsMap)
	return &types.PreNewOrderResp{Code: "10000", Msg: "结算完成，请下订单", Data: &types.PreNewOrderRp{PreOrderInfo: orderinfo}}, nil
}
func (l *PreneworderLogic) order2orderInfo(req *types.PreNewOrderRes, productsMap map[int64]*types.ProductInfo) (orderinfo *types.PreOrderInfo) {

	orderinfo = &types.PreOrderInfo{}
	orderinfo.Phone = req.Phone

	for _, tiny := range req.ProductTinyList {
		infopt, ok := productsMap[tiny.PId]
		if !ok {
			fmt.Println(infopt)
			continue
		}
		orderinfo.OriginalAmount = orderinfo.OriginalAmount + productsMap[tiny.PId].Promotion_price*float64(tiny.Amount)
	}
	l.calculatemoney(req.UsedCouponId, req.UseCouponFirst, req.UseCashFirst, req.Phone, orderinfo)
	orderinfo.FreightAmount = 40 // 后面要增加运费生成模块
	orderinfo.PidList = req.ProductTinyList
	orderinfo.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	return orderinfo
}

func getsha256(msg string) string {
	bytes := sha256.Sum256([]byte(msg))       //计算哈希值，返回一个长度为32的数组
	hashCode2 := hex.EncodeToString(bytes[:]) //将数组转换成切片，转换成16进制，返回字符串
	return hashCode2
}

func (l *PreneworderLogic) calculatemoney(couponid int64, usecoupon, usecash bool, phone string, orderinfo *types.PreOrderInfo) *types.PreOrderInfo {
	//计算打折后的钱
	if usecoupon {
		couponinfo, _ := l.svcCtx.Coupon.FindOneByCouponId(l.ctx, couponid)
		if couponinfo == nil {
			orderinfo.ActualAmount = orderinfo.OriginalAmount
		} else {
			if couponinfo.Discount != 0 {
				orderinfo.ActualAmount = orderinfo.OriginalAmount * float64(couponinfo.Discount) / 100

			} else if couponinfo.MinPoint != 0 && couponinfo.Cut != 0 {
				if orderinfo.ActualAmount < float64(couponinfo.MinPoint) {
					orderinfo.ActualAmount = orderinfo.OriginalAmount
				} else {
					orderinfo.ActualAmount = orderinfo.OriginalAmount - float64(int(orderinfo.OriginalAmount/float64(couponinfo.MinPoint)))
				}
			} else {
				orderinfo.ActualAmount = orderinfo.OriginalAmount
			}
		}
		orderinfo.CouponAmount = orderinfo.OriginalAmount - orderinfo.ActualAmount

	} else {
		orderinfo.ActualAmount = orderinfo.OriginalAmount
		orderinfo.CouponAmount = 0
	}

	// usecash
	if usecash {
		cash, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, phone)
		if cash != nil {
			if (orderinfo.ActualAmount - float64(cash.Balance)) >= 0 {
				orderinfo.WeXinPayAmount = orderinfo.ActualAmount - float64(cash.Balance)
				orderinfo.CashAccountPayAmount = float64(cash.Balance)
			} else {
				orderinfo.WeXinPayAmount = 0
				orderinfo.CashAccountPayAmount = orderinfo.ActualAmount
			}

		} else {
			orderinfo.WeXinPayAmount = orderinfo.ActualAmount
		}
	} else {
		orderinfo.WeXinPayAmount = orderinfo.ActualAmount
		orderinfo.CashAccountPayAmount = 0
	}

	return orderinfo
}
