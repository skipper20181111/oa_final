package qrcode

import (
	"context"
	"encoding/json"
	"math/rand"
	"oa_final/internal/logic/coupon"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetqrcodetypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetqrcodetypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetqrcodetypeLogic {
	return &GetqrcodetypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetqrcodetypeLogic) Getqrcodetype(req *types.ScanQRcodeRes) (resp *types.ScanQRcodeResp, err error) {
	msg := &types.ScanQRcodeRp{Type: "unknow", Msg: randStr(10)}
	//voucher券的parameter1是vid，old2new的券是老用户手机号
	decrypt, err := coupon.Decrypt(req.QRcodeMsg, svc.Keystr)
	if err != nil || decrypt == "" {
		return &types.ScanQRcodeResp{Code: "10000", Msg: "请检查二维码或1分钟后再试", Data: msg}, nil
	}
	QRCodeMsg := &types.QrCode{}
	err = json.Unmarshal([]byte(decrypt), QRCodeMsg)
	if err != nil {
		return &types.ScanQRcodeResp{Code: "10000", Msg: "请检查二维码或1分钟后再试", Data: msg}, nil
	}
	switch QRCodeMsg.Type {
	case "voucher":
		msg = &types.ScanQRcodeRp{Type: "voucher", Msg: randStr(10)}
		return &types.ScanQRcodeResp{Code: "10000", Msg: "success", Data: msg}, nil
	}
	return &types.ScanQRcodeResp{Code: "10000", Msg: "请检查二维码", Data: msg}, nil
}

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
