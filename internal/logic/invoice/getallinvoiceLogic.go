package invoice

import (
	"context"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetallinvoiceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetallinvoiceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetallinvoiceLogic {
	return &GetallinvoiceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetallinvoiceLogic) Getallinvoice(req *types.GetAllInvoiceRes) (resp *types.GetAllInvoiceResp, err error) {
	phone, _ := l.svcCtx.Invoice.FindAll(l.ctx, l.ctx.Value("phone").(string))
	infos := make([]*types.InvoiceRp, 0)
	if phone != nil || len(phone) >= 1 {
		for _, invoice := range phone {
			infos = append(infos, db2info(invoice))
		}
	}
	return &types.GetAllInvoiceResp{Code: "10000", Msg: "success", Data: infos}, nil
}
