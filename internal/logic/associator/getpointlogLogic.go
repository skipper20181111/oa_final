package associator

import (
	"context"
	"oa_final/cachemodel"
	"oa_final/internal/logic/userorder"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetpointlogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ou     *userorder.Logic
}

func NewGetpointlogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetpointlogLogic {
	return &GetpointlogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ou:     userorder.NewLogic(ctx, svcCtx),
	}
}

func (l *GetpointlogLogic) Getpointlog(req *types.GetPointLogRes) (resp *types.GetPointLogResp, err error) {
	accphone := l.ctx.Value("phone").(string)
	canpost := l.ou.PostLimit(accphone+"cashlog", 6)
	if !canpost {
		return &types.GetPointLogResp{Code: "10000", Msg: "success", Data: make(map[string][]*types.PointLogInfo)}, nil
	}
	limit15, err := l.svcCtx.PointLog.FindAllByPhone(context.Background(), accphone)
	cashlogmap := make(map[string][]*types.PointLogInfo)
	cashlogmap["兑换"] = make([]*types.PointLogInfo, 0)
	cashlogmap["获取"] = make([]*types.PointLogInfo, 0)
	cashlogmap["全部"] = make([]*types.PointLogInfo, 0)
	if limit15 == nil || len(limit15) <= 1 {
		return &types.GetPointLogResp{Code: "10000", Msg: "success", Data: cashlogmap}, nil
	}
	for _, cashLog := range limit15 {
		cashlogmap["全部"] = append(cashlogmap["全部"], db2info(cashLog))
		switch cashLog.Behavior {
		case "兑换":
			cashlogmap["兑换"] = append(cashlogmap["兑换"], db2info(cashLog))
		case "获取":
			cashlogmap["获取"] = append(cashlogmap["获取"], db2info(cashLog))
		}
	}
	return &types.GetPointLogResp{Code: "10000", Msg: "success", Data: cashlogmap}, nil
}
func db2info(db *cachemodel.PointLog) *types.PointLogInfo {
	info := &types.PointLogInfo{}
	info.Phone = db.Phone
	info.ChangeAmount = float64(db.ChangeAmount) / 100
	info.Behavior = db.OrderType
	info.OrderType = db.Behavior
	info.OrderSn = db.OrderSn
	info.OrderDescribe = db.OrderDescribe
	info.LogDate = db.Date.Format("2006-01-02 15:04:05")
	info.Balance = float64(db.Balance) / 100
	return info
}
