package payrecall

import (
	"context"
	"fmt"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TellmesoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTellmesoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TellmesoLogic {
	return &TellmesoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TellmesoLogic) Tellmeso(req *types.TellMeSoRes) (resp *types.TellMeSoResp, err error) {
	resourceinfo := req.Resource
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@   req:", req, "%%%%%%%%resourceinfo:", *resourceinfo)
	return &types.TellMeSoResp{Code: "SUCCESS", Message: "成功"}, nil
}
