package invoice

import (
	"context"
	"encoding/json"
	"oa_final/cachemodel"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatetitleLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	userphone  string
	useropenid string
}

func NewUpdatetitleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatetitleLogic {
	return &UpdatetitleLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		userphone:  ctx.Value("phone").(string),
		useropenid: ctx.Value("openid").(string),
	}
}

func (l *UpdatetitleLogic) Updatetitle(req *types.UpdateTitleRes) (resp *types.GetTitleResp, err error) {
	titleinfo := &types.GetTitleRp{Personal: make([]*types.PersonalInvoiceInfo, 0), Company: make([]*types.CompanyInvoiceInfo, 0)}
	marshal, _ := json.Marshal(req.TitleInfoList)
	phone, _ := l.svcCtx.UserInvoiceString.FindOneByPhone(l.ctx, l.userphone)
	if phone != nil {
		phone.InvoiceString = string(marshal)
		l.svcCtx.UserInvoiceString.Update(l.ctx, phone)
	} else {
		l.svcCtx.UserInvoiceString.Insert(l.ctx, &cachemodel.UserInvoiceString{Phone: l.userphone, InvoiceString: string(marshal)})
	}
	byPhone, _ := l.svcCtx.UserInvoiceString.FindOneByPhone(l.ctx, l.userphone)
	if byPhone != nil {
		json.Unmarshal([]byte(byPhone.InvoiceString), titleinfo)
	}
	return &types.GetTitleResp{Code: "10000", Msg: "success", TitleInfoList: titleinfo}, nil
}
