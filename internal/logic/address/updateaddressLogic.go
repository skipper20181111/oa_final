package address

import (
	"context"
	"encoding/json"
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
	if l.ctx.Value("openid") != req.OpenId || l.ctx.Value("phone") != req.Phone {
		return &types.UpdateAddressResp{
			Code: "4004",
			Msg:  "请勿使用其他用户的token",
		}, nil
	}
	if len(req.AddressInfoList) != 0 {
		count := 0
		havedefalt := 0
		for _, info := range req.AddressInfoList {
			if havedefalt == 1 {
				info.IsDefault = 0
			} else {
				if info.IsDefault == 1 {
					count += 1
					havedefalt = 1
				}
			}
		}
		if count == 0 {
			req.AddressInfoList[0].IsDefault = 1
		}
	}
	//以上这一段是为了修正只能有一个默认地址
	marshaledList, err := json.Marshal(req.AddressInfoList)
	if utf8.RuneCountInString(string(marshaledList)) > 20000 {
		return &types.UpdateAddressResp{Code: "4004", Msg: "超长"}, nil
	}
	findAddressListByPhone, err := l.svcCtx.UserAddressString.FindOneByPhone(l.ctx, req.Phone)
	if findAddressListByPhone == nil && err.Error() == "notfind" {
		l.svcCtx.UserAddressString.Insert(l.ctx, &cachemodel.UserAddressString{Phone: req.Phone, AddressString: string(marshaledList)})
	} else if findAddressListByPhone == nil {
		return &types.UpdateAddressResp{Code: "4004", Msg: "猜测是网络问题"}, nil
	} else {
		l.svcCtx.UserAddressString.UpdateByPhone(l.ctx, &cachemodel.UserAddressString{Phone: req.Phone, AddressString: string(marshaledList)})
	}
	findAddressListByPhone, err = l.svcCtx.UserAddressString.FindOneByPhone(l.ctx, req.Phone)
	addressList := make([]*types.AddressInfo, 0)
	json.Unmarshal([]byte(findAddressListByPhone.AddressString), &addressList)
	return &types.UpdateAddressResp{Code: "10000", Msg: "success", Data: &types.UpdateAddressRp{Address: addressList}}, nil
}
