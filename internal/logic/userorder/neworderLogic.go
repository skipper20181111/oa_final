package userorder

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"oa_final/cachemodel"
	"oa_final/internal/logic/refresh"
	"time"

	"oa_final/internal/svc"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type NeworderLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	cashaccount *cachemodel.CashAccount
	userorder   *cachemodel.UserOrder
	usecash     bool
	usecoupon   bool
	userphone   string
}

func NewNeworderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NeworderLogic {
	return &NeworderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NeworderLogic) Neworder(req *types.NewOrderRes) (resp *types.NewOrderResp, err error) {
	l.userphone = l.ctx.Value("phone").(string)
	useropenid := l.ctx.Value("openid").(string)
	lid := time.Now().UnixNano() + int64(rand.Intn(1024))
	l.oplog("付款啊", l.userphone, "开始更新", lid)
	if len(req.ProductTinyList) == 0 {
		return &types.NewOrderResp{Code: "4004", Msg: "无商品，订单金额为0", Data: &types.NewOrderRp{}}, nil
	}
	PMcache, ok := l.svcCtx.LocalCache.Get(refresh.ProductsMap)
	if !ok {
		return &types.NewOrderResp{Code: "4004", Msg: "服务器查找商品列表失败"}, nil
	}
	productsMap := PMcache.(map[int64]*types.ProductInfo)
	order := l.order2db(req, productsMap)
	//从这里开始更新现金账户于优惠券账户
	// 此时还有特别重要的事情，1，要更改现金账户余额，2，要更改优惠券账户，毕竟优惠券账户已经用完了。
	if l.usecash || l.usecoupon {
		lockmsglist := make([]*types.LockMsg, 0)
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.ctx.Value("phone").(string), Field: "user_coupon"})
		lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.ctx.Value("phone").(string), Field: "cash_account"})
		if l.getlock(lockmsglist) {
			if l.usecash {
				if !l.updatecashaccount(lid) {
					order.WexinPayAmount = order.ActualAmount
					order.CashAccountPayAmount = 0
					l.oplog("支付模块更新现金账户失败", l.userphone, "开始更新", lid)
				}
			}

			if l.usecoupon {
				if !l.updatecoupon(lid) {
					l.oplog("支付模块更新优惠券失败", l.userphone, "开始更新", lid)
				}
			}

		}
	}

	//结束更新现金账户与优惠券账户
	l.svcCtx.UserOrder.Insert(l.ctx, order)
	sn2order, err := l.svcCtx.UserOrder.FindOneByOrderSn(l.ctx, order.OrderSn)
	if sn2order == nil {
		fmt.Println(err.Error())
		return &types.NewOrderResp{Code: "4004", Msg: "数据库失效"}, nil
	}
	l.userorder = sn2order
	orderinfo := db2orderinfo(sn2order)
	money := sn2order.WexinPayAmount
	if money == 0 {
		return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &types.NewOrderRp{OrderInfo: orderinfo, UseWechatPay: false}}, nil
	}
	// 此处开始生成订单
	l.oplog("微信支付啊", l.userphone, "开始更新", lid)
	jssvc := jsapi.JsapiApiService{Client: l.svcCtx.Client}
	// 得到prepay_id，以及调起支付所需的参数和签名
	payment, result, err := jssvc.PrepayWithRequestPayment(l.ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(l.svcCtx.Config.WxConf.AppId),
			Mchid:       core.String(l.svcCtx.Config.WxConf.MchID),
			Description: core.String("沾还是不沾芥末，这是一个问题"),
			OutTradeNo:  core.String(orderinfo.OutTradeNo),
			Attach:      core.String(randStr(16)),
			NotifyUrl:   core.String(l.svcCtx.Config.ServerInfo.Url + "/payrecall/tellmeso"),
			Amount: &jsapi.Amount{
				Total: core.Int64(money),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(useropenid),
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
	neworderrp := types.NewOrderRp{OrderInfo: orderinfo, UseWechatPay: true, TimeStamp: timestampsec, NonceStr: nonceStr, Package: packagestr, SignType: signType, PaySign: paySign}
	l.oplog("微信支付啊", l.userphone, "结束更新", lid)
	l.oplog("付款啊", l.userphone, "结束更新", lid)
	return &types.NewOrderResp{Code: "10000", Msg: "success", Data: &neworderrp}, nil

}
func (l *NeworderLogic) updatecashaccount(lid int64) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()

	accphone := l.ctx.Value("phone").(string)
	phone, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, accphone)
	l.oplog("cash_account", accphone, "开始更新", lid)
	phone.Balance = phone.Balance - l.cashaccount.Balance
	l.svcCtx.CashAccount.Update(l.ctx, phone)
	l.oplog("cash_account", accphone, "结束更新", lid)
	return true
}
func (l *NeworderLogic) updatecoupon(lid int64) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	accphone := l.ctx.Value("phone").(string)
	phone, _ := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, accphone)
	l.oplog("usercounpon", accphone, "开始更新", lid)
	usercouponmap := make(map[int64]int)
	json.Unmarshal([]byte(phone.CouponIdList), &usercouponmap)
	usercouponmap[l.userorder.UsedCouponid] = usercouponmap[l.userorder.UsedCouponid] - 1
	marshal, _ := json.Marshal(usercouponmap)
	phone.CouponIdList = string(marshal)
	l.svcCtx.UserCoupon.Update(l.ctx, phone)
	l.oplog("usercounpon", accphone, "结束更新", lid)
	return true
}
func (l *NeworderLogic) oplog(tablename, event, describe string, lid int64) error {
	aol := &cachemodel.AccountOperateLog{Phone: l.ctx.Value("phone").(string), TableName: tablename, Event: event, Describe: describe, Timestamp: time.Now(), Lid: lid}
	_, err := l.svcCtx.AccountOperateLog.Insert(l.ctx, aol)
	return err
}
func (l *NeworderLogic) getlock(lockmsglist []*types.LockMsg) bool {
	//phone := l.ctx.Value("phone").(string)
	lockhost := l.svcCtx.Config.Lock.Host
	urlPath := fmt.Sprintf("%s%s%s", "http://", lockhost, "/pcc/getlock")

	res := types.GetLockRes{LockMsgList: lockmsglist}

	resp, err := httpc.Do(context.Background(), http.MethodPost, urlPath, res)
	if err != nil {

		fmt.Println(err)
	}
	if resp == nil || resp.Body == nil {
		return false
	}
	lockresult := &types.GetLockResp{Code: make(map[string]bool)}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, lockresult)
	defer resp.Body.Close()
	for _, b := range lockresult.Code {
		if b == false {
			return false
		}
	}
	return true

}
func (l *NeworderLogic) closelock(lockmsglist []*types.LockMsg) bool {
	//phone := l.ctx.Value("phone").(string)
	lockhost := l.svcCtx.Config.Lock.Host
	urlPath := fmt.Sprintf("%s%s%s", "http://", lockhost, "/pcc/closelock")
	res := types.GetLockRes{LockMsgList: lockmsglist}

	resp, err := httpc.Do(context.Background(), http.MethodPost, urlPath, res)
	if err != nil {
		fmt.Println(err)
	}
	if resp == nil || resp.Body == nil {
		return false
	}
	lockresult := &types.GetLockResp{Code: make(map[string]bool)}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, lockresult)
	defer resp.Body.Close()
	for _, b := range lockresult.Code {
		if b == false {
			return false
		}
	}
	return true

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
	orderinfo.OriginalAmount = float64(order.OriginalAmount) / 100
	orderinfo.ActualAmount = float64(order.ActualAmount) / 100
	orderinfo.CouponAmount = float64(order.CouponAmount) / 100
	orderinfo.WeXinPayAmount = float64(order.WexinPayAmount) / 100
	orderinfo.CashAccountPayAmount = float64(order.CashAccountPayAmount) / 100
	orderinfo.FreightAmount = float64(order.FreightAmount) / 100
	orderinfo.Growth = order.Growth
	orderinfo.BillType = order.BillType
	orderinfo.BillInfo = &types.Billinfo{}
	orderinfo.OrderStatus = order.OrderStatus
	orderinfo.DeliveryCompany = order.DeliveryCompany
	orderinfo.DeliverySn = order.DeliverySn
	orderinfo.AutoConfirmDay = order.AutoConfirmDay
	address := types.AddressInfo{}
	json.Unmarshal([]byte(order.Address), &address)
	orderinfo.Address = &address
	orderinfo.OrderNote = order.OrderNote
	orderinfo.ConfirmStatus = order.ConfirmStatus
	orderinfo.DeleteStatus = order.DeleteStatus
	orderinfo.PaymentTime = order.PaymentTime.Format("2006-01-02 15:04:05")
	orderinfo.ModifyTime = order.ModifyTime.Format("2006-01-02 15:04:05")
	return orderinfo
}
func (l *NeworderLogic) order2db(req *types.NewOrderRes, productsMap map[int64]*types.ProductInfo) *cachemodel.UserOrder {
	inittime, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	order := &cachemodel.UserOrder{}
	order.Phone = l.userphone
	order.CreateOrderTime = time.Now()
	order.OutTradeNo = randStr(32)
	marshal, err := json.Marshal(req.ProductTinyList)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	order.Pidlist = string(marshal)
	for _, tiny := range req.ProductTinyList {
		order.OriginalAmount = order.OriginalAmount + int64(productsMap[tiny.PId].Promotion_price*100*float64(tiny.Amount))
	}
	l.calculatemoney(req.UsedCouponId, req.UseCouponFirst, req.UseCashFirst, l.userphone, order)
	order.FreightAmount = 4000
	order.OrderStatus = 0
	order.DeliveryCompany = "顺丰"
	order.DeliverySn = randStr(20)
	addr, err := json.Marshal(req.Address)
	if err != nil {
		fmt.Println(err.Error(), "结构体转化为字符串失败")
	}
	order.Address = string(addr)
	//order.BillType = 0
	//order.BillInfo =""

	order.OrderNote = req.OrderNote
	order.DeleteStatus = 0
	order.Growth = order.ActualAmount
	order.ConfirmStatus = 0
	order.ModifyTime = order.CreateOrderTime
	order.PaymentTime = inittime
	order.DeliveryTime = inittime
	order.ReceiveTime = inittime
	order.CloseTime = inittime
	order.OrderSn = getsha512(order.Phone + order.CreateOrderTime.String() + order.Pidlist + order.Address)
	return order
}
func getsha512(message string) string {
	bytes2 := sha512.Sum512([]byte(message))   //计算哈希值，返回一个长度为32的数组
	hashCode2 := hex.EncodeToString(bytes2[:]) //将数组转换成切片，转换成16进制，返回字符串
	return hashCode2
}

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (l *NeworderLogic) calculatemoney(couponid int64, UseCoupon, usecash bool, phone string, orderinfo *cachemodel.UserOrder) *cachemodel.UserOrder {
	l.usecoupon = false
	l.usecash = false
	if UseCoupon {
		//计算打折后的钱
		orderinfo.UsedCouponid = -1
		couponinfo, _ := l.svcCtx.Coupon.FindOneByCouponId(l.ctx, couponid)
		byPhone, _ := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, phone)
		if couponinfo == nil || byPhone == nil {

			orderinfo.ActualAmount = orderinfo.OriginalAmount
		} else {
			usercouponmap := make(map[int64]int)
			json.Unmarshal([]byte(byPhone.CouponIdList), &usercouponmap)
			couponcount := usercouponmap[couponid]
			if couponcount > 0 {
				if couponinfo.Discount != 0 {
					l.usecoupon = true
					orderinfo.UsedCouponid = couponid
					orderinfo.ActualAmount = orderinfo.OriginalAmount * (couponinfo.Discount) / 100

				} else if couponinfo.MinPoint != 0 && couponinfo.Cut != 0 {
					if orderinfo.ActualAmount < couponinfo.MinPoint*100 {
						orderinfo.ActualAmount = orderinfo.OriginalAmount
					} else {
						l.usecoupon = true
						orderinfo.UsedCouponid = couponid
						orderinfo.ActualAmount = orderinfo.OriginalAmount - orderinfo.OriginalAmount/(couponinfo.MinPoint*100)
					}
				} else {
					orderinfo.ActualAmount = orderinfo.OriginalAmount
				}
			} else {
				orderinfo.ActualAmount = orderinfo.OriginalAmount
			}

		}
		orderinfo.CouponAmount = orderinfo.OriginalAmount - orderinfo.ActualAmount

	} else {
		orderinfo.CouponAmount = 0
		orderinfo.ActualAmount = orderinfo.OriginalAmount
	}

	// usecash
	if usecash {
		cash, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, phone)
		if cash != nil {
			if cash.Balance*100 > 0 {
				l.usecash = true
				if (orderinfo.ActualAmount - int64(cash.Balance*100)) >= 0 {
					orderinfo.WexinPayAmount = orderinfo.ActualAmount - int64(cash.Balance*100)
					orderinfo.CashAccountPayAmount = int64(cash.Balance * 100)
				} else {
					orderinfo.WexinPayAmount = 0
					orderinfo.CashAccountPayAmount = orderinfo.ActualAmount
				}
			} else {
				orderinfo.WexinPayAmount = orderinfo.ActualAmount
				orderinfo.CashAccountPayAmount = 0
			}

		} else {
			orderinfo.WexinPayAmount = orderinfo.ActualAmount
		}
	} else {
		orderinfo.WexinPayAmount = orderinfo.ActualAmount
	}

	return orderinfo
}
