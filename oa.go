package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"math/rand"
	"net/http"
	"oa_final/cachemodel"
	"oa_final/internal/logic/orderpay"
	"oa_final/internal/types"
	"time"

	"oa_final/internal/config"
	"oa_final/internal/handler"
	"oa_final/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/oa-api.yaml", "the config file")

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	go refresscache()
	go PrepareGoods(ctx)
	go monitorOrder(ctx)
	go IfReceived(ctx)
	go delivering(ctx)
	go wxnmsl(ctx)
	server.Start()
}
func delivering(ctx *svc.ServiceContext) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	for true {
		RefreshGap := time.Minute * time.Duration(rand.Intn(30)+1)
		time.Sleep(RefreshGap)
		//time.Sleep(time.Second)
		backcontext := context.Background()
		orderlist, _ := ctx.Order.FindDelivering(backcontext)
		if orderlist != nil && len(orderlist) > 0 {
			sf := orderpay.NewSfUtilLogic(backcontext, ctx)
			for _, order := range orderlist {
				sf.IfDelivering(order)
			}
		}
	}
}
func IfReceived(ctx *svc.ServiceContext) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	for true {
		RefreshGap := time.Minute * time.Duration(rand.Intn(120)+1)
		time.Sleep(RefreshGap)
		//time.Sleep(time.Second * 10)
		backcontext := context.Background()
		backcontext = context.WithValue(backcontext, "phone", "17854230845")
		backcontext = context.WithValue(backcontext, "openid", "17854230845")
		orderlist, _ := ctx.Order.FindStatus2(backcontext)
		if orderlist != nil && len(orderlist) > 0 {
			sf := orderpay.NewSfUtilLogic(backcontext, ctx)
			for _, order := range orderlist {
				sf.IfReceived(order)
			}
		}
		//OutTradeSnList, _ := ctx.Order.FindStatus3(backcontext)
		//for _, OutTradeSn := range OutTradeSnList {
		//	status, _ := ctx.Order.FindAllStatusByOutTradeNo(backcontext, OutTradeSn)
		//	if len(status) == 1 && status[0] == 3 {
		//		payInfo, _ := ctx.PayInfo.FindOneByOutTradeNo(backcontext, OutTradeSn)
		//		if payInfo != nil {
		//			ctx.PayInfo.UpdateStatus(backcontext, OutTradeSn, 4)
		//			ctx.Order.UpdateClosedByOutTradeSn(backcontext, OutTradeSn)
		//			ctx.UserPoints.UpdatePoints(backcontext, payInfo.Phone, payInfo.TotleAmount)
		//			userPoints, _ := ctx.UserPoints.FindOneByPhone(backcontext, payInfo.Phone)
		//			ctx.PointLog.Insert(backcontext, &cachemodel.PointLog{Date: time.Now(),
		//				OrderType:     "正常商品",
		//				OrderSn:       payInfo.OutTradeNo,
		//				OrderDescribe: "正常商品收货获取积分",
		//				Behavior:      "获取",
		//				Phone:         payInfo.Phone,
		//				Balance:       userPoints.AvailablePoints,
		//				ChangeAmount:  payInfo.TotleAmount/100 + 1,
		//			})
		//		}
		//	}
		//}
	}
}
func PrepareGoods(SvcCtx *svc.ServiceContext) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	for true {
		RefreshGap := time.Second * time.Duration(rand.Intn(10)+1)
		time.Sleep(RefreshGap)
		//time.Sleep(time.Second * 10)
		ctx := context.Background()
		Orders, _ := SvcCtx.Order.FindStatusBiggerThan1(ctx)
		if len(Orders) > 0 {
			sf := orderpay.NewSfUtilLogic(context.Background(), SvcCtx)
			for _, order := range Orders {
				sf.GetSfSn(order)
			}
		}
	}
}

func monitorOrder(ctx *svc.ServiceContext) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	for true {
		RefreshGap := time.Second * time.Duration(rand.Intn(100)+50)
		time.Sleep(RefreshGap)
		//time.Sleep(time.Second * 10)
		backcontext := context.Background()
		backcontext = context.WithValue(backcontext, "phone", "17854230845")
		backcontext = context.WithValue(backcontext, "openid", "17854230845")
		changed, _ := ctx.Order.FindCanChanged(backcontext)
		if changed != nil && len(changed) > 0 {
			l := orderpay.NewCheckOrderLogic(backcontext, ctx)
			for _, order := range changed {
				if order.OrderStatus == 0 || order.OrderStatus == 6 {
					payinfo, _ := ctx.PayInfo.FindOneByOutTradeNo(backcontext, order.OutTradeNo)
					if !orderpay.PartPay(payinfo) && order.CreateOrderTime.Add(time.Minute*15).Before(time.Now()) {
						ctx.Order.UpdateStatusByOrderSn(backcontext, 8, order.OrderSn)
					} else {
						l.Checkall(order, payinfo)
					}
				}
			}
		}
		PayInfos, _ := ctx.PayInfo.FindStatus0(backcontext)
		for _, PayInfo := range PayInfos {
			deleted, _ := ctx.Order.FindAllByOutTradeNoNotDeleted(backcontext, PayInfo.OutTradeNo)
			if len(deleted) == 0 {
				ctx.PayInfo.UpdateStatus(backcontext, PayInfo.OutTradeNo, 8)
			}
		}

	}
}
func refresscache() {
	for true {
		fmt.Println("开始刷新")
		time.Sleep(time.Second * 1)
		urlPath := "http://localhost:8888/refresh/refreshPL"
		resp, _ := httpc.Do(context.Background(), http.MethodGet, urlPath, nil)
		if resp == nil {
			time.Sleep(time.Second * 50)
			continue
		} else {
			fmt.Println("结束刷新", resp)
			fmt.Println(resp.Body.Close())
			time.Sleep(time.Second * 50)
		}
	}
}

func wxnmsl(svcCtx *svc.ServiceContext) {
	RefreshGapOre := time.Second * time.Duration(rand.Intn(50)+10)
	time.Sleep(RefreshGapOre)
	OuTradeSnListHistory, _ := svcCtx.Order.FindDeliveredOuTradeSnHistory(context.Background())
	for _, OutTradeSn := range OuTradeSnListHistory {
		PayInfo, _ := svcCtx.PayInfo.FindOneByOutTradeNo(context.Background(), OutTradeSn)
		giveMHTshit(svcCtx, PayInfo)
		ConfirmMHTshit(svcCtx, PayInfo)
	}
	for true {
		RefreshGap := time.Minute * time.Duration(rand.Intn(7)+1)
		time.Sleep(RefreshGap)
		endtimestamp := time.Now()
		starttimestamp := endtimestamp.Add(time.Minute * (-10))
		OuTradeSnList, _ := svcCtx.Order.FindDeliveredOuTradeSn(context.Background(), starttimestamp, endtimestamp)
		for _, OutTradeSn := range OuTradeSnList {
			PayInfo, _ := svcCtx.PayInfo.FindOneByOutTradeNo(context.Background(), OutTradeSn)
			giveMHTshit(svcCtx, PayInfo)
			ConfirmMHTshit(svcCtx, PayInfo)
		}
	}
}
func ConfirmMHTshit(svcCtx *svc.ServiceContext, PayInfo *cachemodel.PayInfo) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	ctx := context.Background()
	accessToken, _ := svcCtx.AccessToken.FindOne(ctx, 1)
	UrlPath := fmt.Sprintf("https://api.weixin.qq.com/wxa/sec/order/get_order?access_token=%s", accessToken.Token)
	resp, _ := httpc.Do(context.Background(), http.MethodPost, UrlPath, types.MsgDelivering{TransactionId: PayInfo.TransactionId})
	res := types.MsgReturn{}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.Body.Close())
	json.Unmarshal(body, &res)
	switch res.Order.OrderState {
	case 2:
		svcCtx.PayInfo.UpdateWeChatDelivering(ctx, PayInfo.OutTradeNo)
		svcCtx.Order.UpdateWeChatDeliveredByOutTradeSn(ctx, PayInfo.OutTradeNo)
	case 3, 4:
		svcCtx.PayInfo.UpdateWeChatDelivered(ctx, PayInfo.OutTradeNo)
		svcCtx.Order.UpdateWeChatDeliveredByOutTradeSn(ctx, PayInfo.OutTradeNo)
	}
}
func giveMHTshit(svcCtx *svc.ServiceContext, PayInfo *cachemodel.PayInfo) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	ctx := context.Background()

	orders, _ := svcCtx.Order.FindAllByOutTradeNo(ctx, PayInfo.OutTradeNo)
	userinfos, _ := svcCtx.UserModel.FindOneByPhone(ctx, PayInfo.Phone)
	if len(orders) > 10 {
		orders = orders[:10]
	}
	shippinginfo := make([]*types.ShippingList, 0)
	for _, order := range orders {
		switch order.OrderStatus {
		case 1001, 1002, 1003, 2, 3, 4:
			shippinginfo = append(shippinginfo, &types.ShippingList{
				TrackingNo:     order.DeliverySn,
				ExpressCompany: "SF",
				ItemDesc:       order.ProductInfo,
				Contact: &types.Contact{
					ConsignorContact: "178****0845",
				},
			})
		}

	}
	DataMsg := types.MsgData{
		OrderKey: &types.OrderKey{
			OrderNumberType: 2,
			TransactionId:   PayInfo.TransactionId,
			//Mchid:           "1652716843",
			//OutTradeNo:      "lt9i1DBXoZiefJD5pfPyS8zMV1P2i7GL",
		},
		LogisticsType:  1,
		DeliveryMode:   2,
		IsAllDelivered: true,
		ShippingList:   shippinginfo,
		UploadTime:     time.Now().Format("2006-01-02T15:04:05.000+08:00"),
		Payer: &types.Payer{
			Openid: userinfos.Openid,
		},
	}
	accessToken, _ := svcCtx.AccessToken.FindOne(ctx, 1)
	UrlPath := fmt.Sprintf("https://api.weixin.qq.com/wxa/sec/order/upload_shipping_info?access_token=%s", accessToken.Token)
	resp, _ := httpc.Do(context.Background(), http.MethodPost, UrlPath, DataMsg)
	fmt.Println(resp.Body.Close())

}
