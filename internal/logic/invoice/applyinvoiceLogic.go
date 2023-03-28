package invoice

import (
	"context"
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
}

func NewApplyinvoiceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyinvoiceLogic {
	return &ApplyinvoiceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyinvoiceLogic) Applyinvoice(req *types.ApplyInvoiceRes) (resp *types.ApplyInvoiceResp, err error) {
	l.userphone = l.ctx.Value("phone").(string)
	sn, _ := l.svcCtx.Invoice.FindOneByOrderSn(l.ctx, req.OrderSn)
	orderSn := &cachemodel.Invoice{}
	if sn != nil {
		if sn.Status == 1 {
			return &types.ApplyInvoiceResp{Code: "10000", Msg: "此订单已经开过发票"}, nil
		} else {
			l.svcCtx.Invoice.Update(l.ctx, l.req2db(req))
			orderSn, _ = l.svcCtx.Invoice.FindOneByOrderSn(l.ctx, req.OrderSn)
		}
	} else {
		l.svcCtx.Invoice.Insert(l.ctx, l.req2db(req))
		orderSn, _ = l.svcCtx.Invoice.FindOneByOrderSn(l.ctx, req.OrderSn)
	}
	if orderSn.OrderSn == req.OrderSn {
		return &types.ApplyInvoiceResp{Code: "10000", Msg: "success", Data: db2info(orderSn)}, nil
	} else {
		return &types.ApplyInvoiceResp{Code: "10000", Msg: "数据库失效"}, nil
	}
}
func db2info(db *cachemodel.Invoice) *types.InvoiceInfo {
	info := &types.InvoiceInfo{}
	info.TaxId = db.TaxId
	info.InvoiceTitle = db.InvoiceTitle
	info.OpeningBank = db.OpeningBank
	info.BankAccount = db.BankAccount
	info.ComponyAddress = db.ComponyAddress
	info.ComponyPhone = db.ComponyPhone
	info.Phone = db.Phone
	info.Status = db.Status
	info.OrderSn = db.OrderSn
	info.OrderType = db.OrderType
	info.InvoiceType = db.Type
	info.ApplyTime = db.ApplyTime.Format("2006-01-02 15:04:05")
	info.FinishTime = db.FinishTime.Format("2006-01-02 15:04:05")

	return info

}
func (l *ApplyinvoiceLogic) req2db(req *types.ApplyInvoiceRes) *cachemodel.Invoice {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	db := &cachemodel.Invoice{}
	db.OrderSn = req.OrderSn
	db.Phone = l.userphone
	db.Status = 0
	db.Type = req.InvoiceType
	db.OrderType = req.OrderType
	db.ApplyTime = time.Now()
	db.FinishTime = inittime
	db.InvoiceTitle = req.InvoiceTitle
	db.ComponyPhone = req.ComponyPhone
	db.ComponyAddress = req.ComponyAddress
	db.OpeningBank = req.OpeningBank
	db.BankAccount = req.BankAccount
	db.TaxId = req.TaxId
	return db
}
