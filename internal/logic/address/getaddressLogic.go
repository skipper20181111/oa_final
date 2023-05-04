package address

import (
	"context"
	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetaddressLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	address *AddressUtileLogic
}

func NewGetaddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetaddressLogic {
	return &GetaddressLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		address: NewAddressUtileLogic(ctx, svcCtx),
	}
}

func (l *GetaddressLogic) Getaddress(req *types.GetAddressRes) (resp *types.GetAddressResp, err error) {
	addressList := l.address.Getaddress()
	if addressList != nil {
		return &types.GetAddressResp{Code: "10000", Msg: "success", Data: &types.GetAddressRp{Address: addressList}}, nil
	}
	return &types.GetAddressResp{Code: "10000", Msg: "success", Data: &types.GetAddressRp{Address: make([]*types.AddressInfo, 0)}}, nil

}
