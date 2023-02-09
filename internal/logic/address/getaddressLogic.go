package address

import (
	"context"
	"encoding/json"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetaddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetaddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetaddressLogic {
	return &GetaddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetaddressLogic) Getaddress(req *types.GetAddressRes) (resp *types.GetAddressResp, err error) {
	userphone := l.ctx.Value("phone").(string)
	findAddressListByPhone, err := l.svcCtx.UserAddressString.FindOneByPhone(l.ctx, userphone)
	if err != nil {
		return &types.GetAddressResp{Code: "10000", Msg: "success", Data: &types.GetAddressRp{Address: make([]*types.AddressInfo, 0)}}, nil

	}
	addressList := make([]*types.AddressInfo, 0)
	json.Unmarshal([]byte(findAddressListByPhone.AddressString), &addressList)
	return &types.GetAddressResp{Code: "10000", Msg: "success", Data: &types.GetAddressRp{Address: addressList}}, nil
}
