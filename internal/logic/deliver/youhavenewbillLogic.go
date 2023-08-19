package deliver

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type YouhavenewbillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewYouhavenewbillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *YouhavenewbillLogic {
	return &YouhavenewbillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *YouhavenewbillLogic) Youhavenewbill(req *types.NewBillRes) (resp *types.NewBillResp, err error) {
	// todo: add your logic here and delete this line

	return
}
