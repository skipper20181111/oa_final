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
	server.Start()
}
func monitorOrder(ctx *svc.ServiceContext) {
	for true {
		time.Sleep(time.Second * 100)
		backcontext := context.Background()
		backcontext = context.WithValue(backcontext, "phone", "17854230845")
		backcontext = context.WithValue(backcontext, "openid", "17854230845")
		changed, _ := ctx.Order.FindCanChanged(backcontext)
		if changed != nil && len(changed) > 0 {
			l := orderpay.NewCheckOrderLogic(backcontext, ctx)
			sf := orderpay.NewSfUtilLogic(backcontext, ctx)
			for _, order := range changed {
				if order.OrderStatus == 0 || order.OrderStatus == 6 {
					payinfo, _ := ctx.PayInfo.FindOneByOutTradeNo(backcontext, order.OutTradeNo)
					if !orderpay.PartPay(payinfo) && order.CreateOrderTime.Add(time.Minute*15).Before(time.Now()) {
						ctx.Order.UpdateStatusByOrderSn(backcontext, 8, order.OrderSn)
					} else {
						l.Checkall(order, payinfo)
					}
				} else if order.OrderStatus == 1 {
					sf.GetSfSn(order)
				} else if order.OrderStatus == 2 {
					sf.IfReceived(order)
				}
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
