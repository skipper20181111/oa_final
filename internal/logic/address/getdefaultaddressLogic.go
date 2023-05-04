package address

import (
	"context"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetdefaultaddressLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	address *AddressUtileLogic
}

func NewGetdefaultaddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetdefaultaddressLogic {
	return &GetdefaultaddressLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		address: NewAddressUtileLogic(ctx, svcCtx),
	}
}

func (l *GetdefaultaddressLogic) Getdefaultaddress(req *types.GetAddressRes) (resp *types.GetDefaultAddressResp, err error) {
	addressList := l.address.Getaddress()
	if addressList != nil {
		if len(addressList) > 0 {
			for _, info := range addressList {
				if info.IsDefault == 1 {
					return &types.GetDefaultAddressResp{Code: "10000", Msg: "success", Data: &types.GetDefaultAddressRp{Address: info}}, nil
				}
			}
		}
	}
	return &types.GetDefaultAddressResp{Code: "10000", Msg: "success", Data: &types.GetDefaultAddressRp{Address: &types.AddressInfo{}}}, nil

}
