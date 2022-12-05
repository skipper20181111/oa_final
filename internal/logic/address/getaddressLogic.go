package address

import (
	"context"
	"encoding/json"
	"fmt"

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
	findAddressListByPhone, err := l.svcCtx.UserAddressString.FindOneByPhone(l.ctx, req.Phone)
	if err != nil {
		fmt.Println("##############", err.Error())
		return &types.GetAddressResp{Code: "4004", Msg: "没有找到数据", Data: &types.GetAddressRp{}}, nil

	}
	addressList := make([]*types.AddressInfo, 0)
	json.Unmarshal([]byte(findAddressListByPhone.AddressString), &addressList)
	return &types.GetAddressResp{Code: "10000", Msg: "success", Data: &types.GetAddressRp{Address: addressList}}, nil
}
