package invoice

import (
	"context"
	"encoding/json"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyinvoiceLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
	order     *cachemodel.Order
}

func NewApplyinvoiceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyinvoiceLogic {
	return &ApplyinvoiceLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		userphone: ctx.Value("phone").(string),
	}
}

func (l *ApplyinvoiceLogic) Applyinvoice(req *types.ApplyInvoiceRes) (resp *types.ApplyInvoiceResp, err error) {
	invoice, _ := l.svcCtx.Invoice.FindOneByOutTradeNo(l.ctx, req.OutTradeSn)
	payInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, req.OutTradeSn)
	if payInfo == nil || payInfo.Status != 4 {
		return &types.ApplyInvoiceResp{Code: "10000", Msg: "此订单号不可开发票（无此订单或订单未完成）"}, nil
	}
	InvoiceByOTS := &cachemodel.Invoice{}
	if invoice != nil {
		if invoice.Status > 1 {
			return &types.ApplyInvoiceResp{Code: "10000", Msg: "此订单已经开过发票或开票失败"}, nil
		} else {
			newsn := l.req2db(req, payInfo)
			newsn.Id = invoice.Id
			l.svcCtx.Invoice.Update(l.ctx, newsn)
		}
	} else {
		l.svcCtx.Invoice.Insert(l.ctx, l.req2db(req, payInfo))
		l.svcCtx.Order.UpdateInvoice(l.ctx, req.OutTradeSn, 1)
	}
	InvoiceByOTS, _ = l.svcCtx.Invoice.FindOneByOutTradeNo(l.ctx, req.OutTradeSn)
	return &types.ApplyInvoiceResp{Code: "10000", Msg: "success", Data: db2info(InvoiceByOTS)}, nil
}
func db2info(db *cachemodel.Invoice) *types.InvoiceRp {
	info := &types.InvoiceRp{PostAddress: &types.AddressInfo{}, InvoinceInfo: &types.InvoiceInfo{}}
	info.InvoinceInfo.InvoiceType = db.InvoiceType
	info.InvoinceInfo.TargetType = db.Target
	info.InvoinceInfo.TaxId = db.TaxId
	info.Amount = db.Money
	info.InvoinceInfo.IfDetail = db.Ifdetail
	info.InvoinceInfo.InvoiceTitle = db.InvoiceTitle
	info.InvoinceInfo.OpeningBank = db.OpeningBank
	info.InvoinceInfo.BankAccount = db.BankAccount
	info.InvoinceInfo.CompanyAddress = db.CompanyAddress
	info.InvoinceInfo.CompanyPhone = db.CompanyPhone

	addressinfo := &types.AddressInfo{}
	json.Unmarshal([]byte(db.PostAddress), addressinfo)
	info.PostAddress = addressinfo

	info.Phone = db.Phone
	info.Status = db.Status
	info.OutTradeSn = db.OutTradeNo
	info.OrderType = db.OrderType
	info.ApplyTime = db.ApplyTime.Format("2006-01-02 15:04:05")
	info.FinishTime = db.FinishTime.Format("2006-01-02 15:04:05")

	return info

}
func (l *ApplyinvoiceLogic) req2db(req *types.ApplyInvoiceRes, payInfo *cachemodel.PayInfo) *cachemodel.Invoice {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "2099-01-01 00:00:00")
	db := &cachemodel.Invoice{}
	db.OutTradeNo = req.OutTradeSn
	db.InvoiceSn = req.OutTradeSn
	db.Phone = l.userphone
	db.OrderType = req.OrderType
	db.Ifdetail = req.InvoinceInfo.IfDetail
	db.Status = 1
	db.ApplyTime = time.Now()
	db.FinishTime = inittime
	postaddress, _ := json.Marshal(req.PostAddress)
	db.PostAddress = string(postaddress)
	db.InvoiceType = req.InvoinceInfo.InvoiceType
	db.Target = req.InvoinceInfo.TargetType
	db.Money = payInfo.WexinPayAmount
	db.InvoiceTitle = req.InvoinceInfo.InvoiceTitle
	db.CompanyAddress = req.InvoinceInfo.CompanyAddress
	db.CompanyPhone = req.InvoinceInfo.CompanyPhone
	db.TaxId = req.InvoinceInfo.TaxId
	db.BankAccount = req.InvoinceInfo.BankAccount
	db.OpeningBank = req.InvoinceInfo.OpeningBank
	db.Email = req.Email
	return db
}
