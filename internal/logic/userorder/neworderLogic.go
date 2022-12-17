package userorder

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"log"
	"math/rand"
	"oa_final/cachemodel"
	"oa_final/internal/logic/refresh"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type NeworderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNeworderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NeworderLogic {
	return &NeworderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NeworderLogic) Neworder(req *types.NewOrderRes) (resp *types.NewOrderResp, err error) {
	PMcache, ok := l.svcCtx.LocalCache.Get(refresh.ProductsMap)
	if !ok {
		return &types.NewOrderResp{Code: "4004", Msg: "服务器查找商品列表失败"}, nil
	}
	productsMap := PMcache.(map[int64]*types.ProductInfo)
	order := order2db(req, productsMap)
	insert, err := l.svcCtx.UserOrder.Insert(l.ctx, order)
	if err != nil { // 如果插入失败，就多试几次，如果试了三次都失败，那我没办法了
		fmt.Println(insert, err.Error())
		order.OutTradeNo = randStr(32)
		insert, err := l.svcCtx.UserOrder.Insert(l.ctx, order)
		if err != nil {
			fmt.Println(insert, err.Error())
			order.OutTradeNo = randStr(32)
			insert, err := l.svcCtx.UserOrder.Insert(l.ctx, order)
			if err != nil {
				fmt.Println(insert, err.Error())
				return &types.NewOrderResp{Code: "4004", Msg: "数据库失效"}, nil
			}
		}

	}
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, order.OrderSn)
	if err != nil {
		fmt.Println(err.Error())
		return &types.NewOrderResp{Code: "4004", Msg: "数据库失效"}, nil
	}
	orderinfo := db2orderinfo(sn2order)

	// 此处开始生成订单
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	// 得到prepay_id，以及调起支付所需的参数和签名
	payment, result, err := jssvc.PrepayWithRequestPayment(l.ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(l.svcCtx.Config.WxConf.AppId),
			Mchid:       core.String(l.svcCtx.Config.WxConf.MchID),
			Description: core.String("沾还是不沾芥末，这是一个问题"),
			OutTradeNo:  core.String(orderinfo.OutTradeNo),
			Attach:      core.String("自定义数据说明" + randStr(5)),
			NotifyUrl:   core.String(l.svcCtx.Config.ServerInfo.Url + "/payrecall/tellmeso"),
			Amount: &jsapi.Amount{
				Total: core.Int64(1),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(req.OpenId),
			},
		},
	)
	defer result.Response.Body.Close()
	if err == nil {
		log.Println(payment, result)
	} else {
		log.Println(err)
		return &types.NewOrderResp{Code: "4004", Msg: err.Error()}, nil
	}
	// 用于返回给前端调起支付的变量与签名串生成器
	timestampsec := *payment.TimeStamp
	nonceStr := *payment.NonceStr
	packagestr := *payment.Package
	paySign := *payment.PaySign
	signType := *payment.SignType
	neworderrp := types.NewOrderRp{OrderInfo: orderinfo, TimeStamp: timestampsec, NonceStr: nonceStr, Package: packagestr, SignType: signType, PaySign: paySign}
	return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &neworderrp}, nil

}
func db2orderinfo(order *cachemodel.UserOrder) *types.OrderInfo {
	orderinfo := &types.OrderInfo{}
	orderinfo.Phone = order.Phone
	orderinfo.OrderSn = order.OrderSn
	orderinfo.OutTradeNo = order.OutTradeNo
	orderinfo.TransactionId = order.TransactionId
	orderinfo.CreateTime = order.CreateOrderTime.Format("2006-01-02 15:04:05")
	pidlist := make([]*types.ProductTiny, 0)
	json.Unmarshal([]byte(order.Pidlist), &pidlist)
	orderinfo.PidList = pidlist
	orderinfo.OriginalAmount = order.OriginalAmount
	orderinfo.PayAmount = order.PayAmount
	orderinfo.FreightAmount = order.FreightAmount
	orderinfo.OrderStatus = order.OrderStatus
	orderinfo.DeliveryCompany = order.DeliveryCompany
	orderinfo.DeliverySn = order.DeliverySn
	orderinfo.AutoConfirmDay = order.AutoConfirmDay
	address := types.AddressInfo{}
	json.Unmarshal([]byte(order.ReceiverInfo), &address)
	orderinfo.Address = &address
	orderinfo.Note = order.Note
	orderinfo.ConfirmStatus = order.ConfirmStatus
	orderinfo.DeleteStatus = order.DeleteStatus
	orderinfo.UseIntegration = order.UseIntegration
	orderinfo.PaymentTime = order.PaymentTime.Format("2006-01-02 15:04:05")
	orderinfo.ModifyTime = order.ModifyTime.Format("2006-01-02 15:04:05")
	return orderinfo
}
func order2db(req *types.NewOrderRes, productsMap map[int64]*types.ProductInfo) *cachemodel.UserOrder {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	order := &cachemodel.UserOrder{}
	order.Phone = req.Phone
	order.CreateOrderTime = time.Now()
	order.OutTradeNo = randStr(32)
	marshal, err := json.Marshal(req.ProductTinyList)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	order.Pidlist = string(marshal)
	for _, tiny := range req.ProductTinyList {
		order.OriginalAmount = order.OriginalAmount + productsMap[tiny.PId].Original_price*float64(tiny.Amount)
		order.PayAmount = order.PayAmount + productsMap[tiny.PId].Promotion_price*float64(tiny.Amount)
	}
	order.FreightAmount = 40
	order.PromotionAmount = 0
	order.IntegrationAmount = 0
	order.CouponAmount = 0
	order.OrderStatus = 0
	order.DeliveryCompany = "顺丰"
	order.DeliverySn = randStr(20)
	addr, err := json.Marshal(req.Address)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	order.ReceiverInfo = string(addr)
	order.Note = req.OrderNote
	order.DeleteStatus = 0
	order.ModifyTime = order.CreateOrderTime
	order.PaymentTime = inittime
	order.DeliveryTime = inittime
	order.ReceiveTime = inittime
	order.CommentTime = inittime
	order.OrderSn = getsha512(order.Phone + order.CreateOrderTime.String() + order.Pidlist + order.ReceiverInfo)
	return order
}
func getsha512(message string) string {
	bytes2 := sha512.Sum512([]byte(message))   //计算哈希值，返回一个长度为32的数组
	hashCode2 := hex.EncodeToString(bytes2[:]) //将数组转换成切片，转换成16进制，返回字符串
	return hashCode2
}

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	rand.Seed(time.Now().Unix())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
