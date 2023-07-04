package associator

import (
	"context"
	"oa_final/cachemodel"
	"oa_final/internal/logic/orderpay"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetpointlogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ou     *orderpay.UtilLogic
}

func NewGetpointlogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetpointlogLogic {
	return &GetpointlogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ou:     orderpay.NewUtilLogic(ctx, svcCtx),
	}
}

func (l *GetpointlogLogic) Getpointlog(req *types.GetPointLogRes) (resp *types.GetPointLogResp, err error) {
	accphone := l.ctx.Value("phone").(string)
	//canpost := l.ou.PostLimit(accphone+"cashlog", 6)
	//if !canpost {
	//	return &types.GetPointLogResp{Code: "10000", Msg: "success", Data: make(map[string][]*types.PointLogInfo)}, nil
	//}
	limit15, err := l.svcCtx.PointLog.FindAllByPhone(context.Background(), accphone)
	cashlogmap := make(map[string][]*types.PointLogInfo)
	cashlogmap["exchange"] = make([]*types.PointLogInfo, 0)
	cashlogmap["achieve"] = make([]*types.PointLogInfo, 0)
	//cashlogmap["全部"] = make([]*types.PointLogInfo, 0)
	if limit15 == nil || len(limit15) <= 1 {
		return &types.GetPointLogResp{Code: "10000", Msg: "success", Data: cashlogmap}, nil
	}
	for _, cashLog := range limit15 {
		//cashlogmap["全部"] = append(cashlogmap["全部"], db2info(cashLog))
		switch cashLog.Behavior {
		case "兑换":
			info := db2info(cashLog)
			info.OrderType = "exchange"
			cashlogmap["exchange"] = append(cashlogmap["exchange"], info)
		case "获取":
			info := db2info(cashLog)
			info.OrderType = "achieve"
			cashlogmap["achieve"] = append(cashlogmap["achieve"], info)
		}
	}
	return &types.GetPointLogResp{Code: "10000", Msg: "success", Data: cashlogmap}, nil
}
func db2info(db *cachemodel.PointLog) *types.PointLogInfo {
	info := &types.PointLogInfo{}
	info.Phone = db.Phone
	info.ChangeAmount = db.ChangeAmount
	info.Behavior = db.OrderType
	info.OrderTypeZh = db.Behavior
	info.OrderSn = db.OrderSn
	info.OrderDescribe = db.OrderDescribe
	info.LogDate = db.Date.Format("2006-01-02 15:04:05")
	info.Balance = db.Balance
	return info
}
