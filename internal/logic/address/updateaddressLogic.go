package address

import (
	"context"
	"encoding/json"
	"fmt"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateaddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateaddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateaddressLogic {
	return &UpdateaddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateaddressLogic) Updateaddress(req *types.UpdateAddressRes) (resp *types.UpdateAddressResp, err error) {
	marshaledList, err := json.Marshal(req.AddressInfoList)
	if utf8.RuneCountInString(string(marshaledList)) > 6144 {
		return &types.UpdateAddressResp{Code: "4004", Msg: "超长"}, nil
	}
	err = l.svcCtx.UserAddressString.UpdateByPhone(l.ctx, &cachemodel.UserAddressString{Phone: req.Phone, AddressString: string(marshaledList)})
	if err != nil {
		insert, err := l.svcCtx.UserAddressString.Insert(l.ctx, &cachemodel.UserAddressString{Phone: req.Phone, AddressString: string(marshaledList)})
		if err != nil {
			fmt.Println(insert, err, "插入失败了，猜测是网络问题")
			return &types.UpdateAddressResp{Code: "10000", Msg: "猜测是网络问题"}, nil

		}
	}
	findAddressListByPhone, err := l.svcCtx.UserAddressString.FindOneByPhone(l.ctx, req.Phone)
	addressList := make([]*types.AddressInfo, 0)
	json.Unmarshal([]byte(findAddressListByPhone.AddressString), &addressList)
	return &types.UpdateAddressResp{Code: "10000", Msg: "success", Data: &types.UpdateAddressRp{Address: addressList}}, nil
}
