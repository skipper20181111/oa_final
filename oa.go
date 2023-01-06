package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpc"
	"net/http"
	"time"

	"oa_final/internal/config"
	"oa_final/internal/handler"
	"oa_final/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/oa-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	go refresscache()
	server.Start()
}
func refresscache() {
	for true {
		fmt.Println("开始刷新")
		time.Sleep(time.Second * 3)
		urlPath := "http://localhost:8888/refresh/refreshPL"
		resp, err := httpc.Do(context.Background(), http.MethodGet, urlPath, nil)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("结束刷新", resp)
		fmt.Println(resp.Body.Close())
		time.Sleep(time.Second * 50)
	}
}
