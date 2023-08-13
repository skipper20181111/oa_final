package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpc"
	"math/rand"
	"net/http"
	"oa_final/internal/logic/orderpay"
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
	go monitorOrder(ctx)
	go IfReceived(ctx)
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
		backcontext := context.Background()
		backcontext = context.WithValue(backcontext, "phone", "17854230845")
		backcontext = context.WithValue(backcontext, "openid", "17854230845")
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
	}
}
func monitorOrder(ctx *svc.ServiceContext) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	for true {
		//RefreshGap := time.Second * time.Duration(rand.Intn(100)+50)
		//time.Sleep(RefreshGap)
		time.Sleep(time.Second * 10)
		backcontext := context.Background()
		backcontext = context.WithValue(backcontext, "phone", "17854230845")
		backcontext = context.WithValue(backcontext, "openid", "17854230845")
		changed, _ := ctx.Order.FindCanChanged(backcontext)
		sf := orderpay.NewSfUtilLogic(context.Background(), ctx)
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
				} else if order.OrderStatus == 1 && len(order.DeliverySn) < 3 {
					sf.GetSfSn(order)
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
