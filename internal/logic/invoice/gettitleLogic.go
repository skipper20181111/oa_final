package invoice

import (
	"context"
	"encoding/json"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GettitleLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	userphone  string
	useropenid string
}

func NewGettitleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GettitleLogic {
	return &GettitleLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
	}
}

func (l *GettitleLogic) Gettitle(req *types.GetTitleRes) (resp *types.GetTitleResp, err error) {
	titleinfo := &types.GetTitleRp{Personal: make([]*types.PersonalInvoiceInfo, 0), Company: make([]*types.CompanyInvoiceInfo, 0)}
	byPhone, _ := l.svcCtx.UserInvoiceString.FindOneByPhone(l.ctx, l.userphone)
	if byPhone != nil {
		json.Unmarshal([]byte(byPhone.InvoiceString), titleinfo)
	}
	return &types.GetTitleResp{Code: "10000", Msg: "success", Data: titleinfo}, nil
}
