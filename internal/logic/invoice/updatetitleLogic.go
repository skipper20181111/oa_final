package invoice

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
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
	req = l.getdefault(req)
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
	return &types.GetTitleResp{Code: "10000", Msg: "success", Data: titleinfo}, nil
}
func (l *UpdatetitleLogic) getdefault(req *types.UpdateTitleRes) *types.UpdateTitleRes {
	if len(req.TitleInfoList.Company) != 0 {
		count := 0
		havedefalt := 0
		defaltindex := -1
		for i, info := range req.TitleInfoList.Company {
			if havedefalt == 1 {
				info.IsDefault = 0
			} else {
				if info.IsDefault == 1 {
					defaltindex = i
					count += 1
					havedefalt = 1
				}
			}
		}
		if count == 0 {
			req.TitleInfoList.Company[0].IsDefault = 1
		} else {
			se := *req.TitleInfoList.Company[defaltindex]
			req.TitleInfoList.Company[defaltindex] = req.TitleInfoList.Company[0]
			req.TitleInfoList.Company[0] = &se
		}
	}
	if len(req.TitleInfoList.Personal) != 0 {
		count := 0
		havedefalt := 0
		defaltindex := -1
		for i, info := range req.TitleInfoList.Personal {
			if havedefalt == 1 {
				info.IsDefault = 0
			} else {
				if info.IsDefault == 1 {
					defaltindex = i
					count += 1
					havedefalt = 1
				}
			}
		}
		if count == 0 {
			req.TitleInfoList.Personal[0].IsDefault = 1
		} else {
			se := *req.TitleInfoList.Personal[defaltindex]
			req.TitleInfoList.Personal[defaltindex] = req.TitleInfoList.Personal[0]
			req.TitleInfoList.Personal[0] = &se
		}
	}
	return req
}
