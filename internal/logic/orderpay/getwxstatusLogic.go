package orderpay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"net/http"
	"oa_final/cachemodel"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetwxstatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetwxstatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetwxstatusLogic {
	return &GetwxstatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetwxstatusLogic) Getwxstatus(req *types.GetWxStatusRes) (resp *types.GetWxStatusResp, err error) {
	payInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, req.OutTradeNo)
	resp = &types.GetWxStatusResp{
		Code: "10000",
		Msg:  "success",
		Data: l.ConfirmMHTshit(payInfo),
	}
	return resp, nil
}
func (l *GetwxstatusLogic) ConfirmMHTshit(Payinfo *cachemodel.PayInfo) *types.MsgReturn {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	ctx := context.Background()
	accessToken, _ := l.svcCtx.AccessToken.FindOne(ctx, 1)
	UrlPath := fmt.Sprintf("https://api.weixin.qq.com/wxa/sec/order/get_order?access_token=%s", accessToken.Token)
	resp, _ := httpc.Do(context.Background(), http.MethodPost, UrlPath, types.MsgDelivering{TransactionId: Payinfo.TransactionId})
	res := types.MsgReturn{}
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json.Unmarshal(body, &res)
	return &res
}
