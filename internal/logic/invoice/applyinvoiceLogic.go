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
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyinvoiceLogic) Applyinvoice(req *types.ApplyInvoiceRes) (resp *types.ApplyInvoiceResp, err error) {
	l.userphone = l.ctx.Value("phone").(string)
	sn, _ := l.svcCtx.Invoice.FindOneByOrderSn(l.ctx, req.OrderSn)
	l.order, _ = l.svcCtx.Order.FindOneByOrderSn(l.ctx, req.OrderSn)
	if l.order == nil || l.order.OrderStatus != 3 {
		return &types.ApplyInvoiceResp{Code: "10000", Msg: "此订单号不可开发票（无此订单或订单未完成）"}, nil
	}
	orderSn := &cachemodel.Invoice{}
	if sn != nil {
		if sn.Status != 0 {
			return &types.ApplyInvoiceResp{Code: "10000", Msg: "此订单已经开过发票或开票失败"}, nil
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
func db2info(db *cachemodel.Invoice) *types.InvoiceRp {
	info := &types.InvoiceRp{}
	info.InvoinceInfo.InvoiceType = db.InvoiceType
	info.InvoinceInfo.TargetType = db.Target
	info.InvoinceInfo.TaxId = db.TaxId
	info.Amount = db.Money
	info.InvoinceInfo.IfDetail = db.Ifdetail
	info.InvoinceInfo.InvoiceTitle = db.InvoiceTitle
	info.InvoinceInfo.OpeningBank = db.OpeningBank
	info.InvoinceInfo.BankAccount = db.BankAccount
	info.InvoinceInfo.ComponyAddress = db.ComponyAddress
	info.InvoinceInfo.ComponyPhone = db.ComponyPhone

	addressinfo := &types.AddressInfo{}
	json.Unmarshal([]byte(db.PostAddress), addressinfo)
	info.PostAddress = addressinfo

	info.Phone = db.Phone
	info.Status = db.Status
	info.OrderSn = db.OrderSn
	info.OrderType = db.OrderType
	info.ApplyTime = db.ApplyTime.Format("2006-01-02 15:04:05")
	info.FinishTime = db.FinishTime.Format("2006-01-02 15:04:05")

	return info

}
func (l *ApplyinvoiceLogic) req2db(req *types.ApplyInvoiceRes) *cachemodel.Invoice {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	db := &cachemodel.Invoice{}
	db.OrderSn = req.OrderSn
	db.Phone = l.userphone
	db.OrderType = req.OrderType
	db.Ifdetail = req.InvoinceInfo.IfDetail
	db.Status = 0
	db.ApplyTime = time.Now()
	db.FinishTime = inittime
	postaddress, _ := json.Marshal(req.PostAddress)
	db.PostAddress = string(postaddress)
	db.InvoiceType = req.InvoinceInfo.InvoiceType
	db.Target = req.InvoinceInfo.TargetType
	db.Money = l.order.WexinPayAmount
	db.InvoiceTitle = req.InvoinceInfo.InvoiceTitle
	db.ComponyAddress = req.InvoinceInfo.ComponyAddress
	db.ComponyPhone = req.InvoinceInfo.ComponyPhone
	db.TaxId = req.InvoinceInfo.TaxId
	db.BankAccount = req.InvoinceInfo.BankAccount
	db.OpeningBank = req.InvoinceInfo.OpeningBank
	return db
}
