package invoice

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetinvoiceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetinvoiceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetinvoiceLogic {
	return &GetinvoiceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetinvoiceLogic) Getinvoice(req *types.GetInvoiceRes) (resp *types.GetInvoiceResp, err error) {
	sn, _ := l.svcCtx.Invoice.FindOneByOrderSn(l.ctx, req.OrderSn)
	if sn != nil && sn.Phone == l.ctx.Value("phone").(string) {
		return &types.GetInvoiceResp{Code: "10000", Msg: "success", Data: db2info(sn)}, nil
	}
	return &types.GetInvoiceResp{Code: "10000", Msg: "success", Data: &types.InvoiceRp{}}, nil
}
