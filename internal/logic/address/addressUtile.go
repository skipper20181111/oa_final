package address

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

type AddressUtileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddressUtileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddressUtileLogic {
	return &AddressUtileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (l *AddressUtileLogic) Getaddress() []*types.AddressInfo {
	userphone := l.ctx.Value("phone").(string)
	findAddressListByPhone, err := l.svcCtx.UserAddressString.FindOneByPhone(l.ctx, userphone)
	if err != nil {
		return nil
	}
	addressList := make([]*types.AddressInfo, 0)
	json.Unmarshal([]byte(findAddressListByPhone.AddressString), &addressList)
	return addressList
}
