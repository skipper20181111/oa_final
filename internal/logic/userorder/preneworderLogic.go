package userorder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	PMcache, ok := l.svcCtx.LocalCache.Get(refresh.ProductsMap)
	if !ok {
		return &types.PreNewOrderResp{Code: "4004", Msg: "服务器查找商品列表失败"}, nil
	}
	productsMap := PMcache.(map[int64]*types.ProductInfo)
	orderinfo := order2orderInfo(req, productsMap)
	return &types.PreNewOrderResp{Code: "10000", Msg: "结算完成，请下订单", Data: &types.PreNewOrderRp{OrderInfo: orderinfo}}, nil
}
func order2orderInfo(req *types.PreNewOrderRes, productsMap map[int64]*types.ProductInfo) (orderinfo *types.OrderInfo) {
	orderinfo = &types.OrderInfo{}
	orderinfo.Phone = req.Phone

	for _, tiny := range req.ProductTinyList {
		orderinfo.OriginalAmount = orderinfo.OriginalAmount + productsMap[tiny.PId].Original_price*float64(tiny.Amount)
		orderinfo.PayAmount = orderinfo.PayAmount + productsMap[tiny.PId].Promotion_price*float64(tiny.Amount)
	}
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
