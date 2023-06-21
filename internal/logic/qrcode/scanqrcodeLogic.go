package qrcode

import (
	"context"
	"encoding/json"
	"oa_final/internal/logic/coupon"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ScanqrcodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	vu     *coupon.VoucherUtileLogic
}

func NewScanqrcodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScanqrcodeLogic {
	return &ScanqrcodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		vu:     coupon.NewVoucherUtileLogic(ctx, svcCtx),
	}
}

func (l *ScanqrcodeLogic) Scanqrcode(req *types.ScanQRcodeRes) (resp *types.ScanQRcodeResp, err error) {
	//voucher券的parameter1是vid，old2new的券是老用户手机号
	decrypt, err := coupon.Decrypt(req.QRcodeMsg, svc.Keystr)
	if err != nil || decrypt == "" {
		return &types.ScanQRcodeResp{Code: "10000", Msg: "请检查二维码或1分钟后再试"}, nil
	}
	QRCodeMsg := &types.QrCode{}
	err = json.Unmarshal([]byte(decrypt), QRCodeMsg)
	if err != nil {
		return &types.ScanQRcodeResp{Code: "10000", Msg: "请检查二维码或1分钟后再试"}, nil
	}
	switch QRCodeMsg.Type {
	case "voucher":
		_, msg := l.vu.VoucherbindByVid(QRCodeMsg)
		return &types.ScanQRcodeResp{Code: "10000", Msg: msg}, nil
	}
	return &types.ScanQRcodeResp{Code: "10000", Msg: "请检查二维码"}, nil
}
