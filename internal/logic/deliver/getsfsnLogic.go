package deliver

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetsfsnLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetsfsnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetsfsnLogic {
	return &GetsfsnLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetsfsnLogic) Getsfsn(req *types.GetSfSnRes) (resp *types.GetSfSnResp, err error) {
	resp = &types.GetSfSnResp{
		Code: "10000",
		Msg:  "success",
		Data: make(map[int64][]string),
	}
	for _, status := range req.StatusList {
		orders, _ := l.svcCtx.Order.FindAllByStatus(l.ctx, status)
		if len(orders) > 0 {
			resp.Data[status] = make([]string, 0)
			for _, order := range orders {
				resp.Data[status] = append(resp.Data[status], order.DeliverySn)
			}
		}
	}
	return
}
