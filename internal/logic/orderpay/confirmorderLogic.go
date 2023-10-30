package orderpay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"net/http"
	"oa_final/cachemodel"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmorderLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	userphone string
	u         *UtilLogic
}

func NewConfirmorderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmorderLogic {
	return &ConfirmorderLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		userphone: ctx.Value("phone").(string),
		u:         NewUtilLogic(ctx, svcCtx),
	}
}

func (l *ConfirmorderLogic) Confirmorder(req *types.ConfirmOrderRes) (resp *types.ConfirmOrderResp, err error) {
	resp = &types.ConfirmOrderResp{
		Code: "10000",
		Msg:  "success",
		Data: &types.GetAllOrderRp{
			OrderInfos: make([]*types.OrderInfo, 0),
		},
	}
	PayInfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, req.OutTradeNo)
	if PayInfo == nil || PayInfo.Status == 4 {
		return resp, nil
	}
	status, _ := l.svcCtx.Order.FindAllStatusByOutTradeNo(l.ctx, req.OutTradeNo)
	yes := true
	for _, sta := range status {
		switch sta {
		case 2, 3, 4:
			continue
		default:
			yes = false
		}
	}
	if yes && (len(PayInfo.TransactionId) < 2 || l.ConfirmMHTshit(PayInfo)) {

		l.svcCtx.PayInfo.UpdateStatus(l.ctx, req.OutTradeNo, 4)
		l.svcCtx.UserPoints.UpdatePoints(l.ctx, PayInfo.Phone, PayInfo.TotleAmount)
		l.svcCtx.Order.UpdateClosedByOutTradeSn(l.ctx, req.OutTradeNo)
		userPoints, _ := l.svcCtx.UserPoints.FindOneByPhone(l.ctx, l.userphone)
		l.svcCtx.PointLog.Insert(l.ctx, &cachemodel.PointLog{Date: time.Now(),
			OrderType:     "正常商品",
			OrderSn:       PayInfo.OutTradeNo,
			OrderDescribe: "正常商品收货获取积分",
			Behavior:      "获取",
			Phone:         l.userphone,
			Balance:       userPoints.AvailablePoints,
			ChangeAmount:  PayInfo.TotleAmount/100 + 1,
		})

	}
	orders, _ := l.svcCtx.Order.FindAllByOutTradeNo(l.ctx, req.OutTradeNo)
	for _, order := range orders {
		resp.Data.OrderInfos = append(resp.Data.OrderInfos, l.u.OrderDb2info(order))
	}
	return resp, nil
}
func (l *ConfirmorderLogic) ConfirmMHTshit(Payinfo *cachemodel.PayInfo) bool {
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
	if len(res.Order.Openid) > 1 && res.Order.OrderState >= 3 {
		l.svcCtx.PayInfo.UpdateWeChatDelivered(l.ctx, Payinfo.OutTradeNo)
		newPayinfo, _ := l.svcCtx.PayInfo.FindOneByOutTradeNo(l.ctx, Payinfo.OutTradeNo)
		if newPayinfo.WexinDeliveryStatus == 3 {
			return true
		}
	}
	return false
}
