package svc

import (
	"context"
	"crypto/rsa"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/cipher/decryptors"
	"github.com/wechatpay-apiv3/wechatpay-go/core/cipher/encryptors"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"log"
	"net/http"
	"net/url"
	"oa_final/cachemodel"
	"oa_final/internal/config"
	"time"
)

const ()
const (
	localCacheExpire   = time.Duration(time.Second * 800)
	localCacheExpire2  = time.Duration(time.Second * 20)
	RechargeProductKey = "RechargeProductKey"
	ProductsMap        = "ProductsMap"
	ProductsInfoMap    = "ProductsInfoMap"
	StarMallMap        = "StarMallMap"
	CouponMapKey       = "CouponMapKey"
	CouponInfoMapKey   = "CouponInfoMapKey"
	Keystr             = "W3WxhhoA9E9VIteCYbnhUTxDbtk2nP1Z"
	MissionListKey     = "MissionListKey"
	//ParterID           = "ZYYMKIJW69LH"
	////测试环境
	//CheckCodeSbox = "v8oHKTufkthd7xCuI9kyX7KDwnGeKFt3"
	//MonthlyCard   = "7551234567"
	//SfUrl         = "https://sfapi-sbox.sf-express.com/std/service"
	////生产环境
	//CheckCodeSbox          = "uB6bnfnBaCxGWtzbQj477KpeAbEOdgjs"
	//MonthlyCard            = "0210927407"
	//SfUrl                  = "https://bspgw.sf-express.com/std/service"
	CreateOrderServiceCode = "EXP_RECE_CREATE_ORDER"
	GetRoutesServiceCode   = "EXP_RECE_SEARCH_ROUTES"
	QueryOrderServiceCode  = "EXP_RECE_SEARCH_ORDER_RESP"
	DownPDFServiceCode     = "COM_RECE_CLOUD_PRINT_WAYBILLS"
	RefundServiceCode      = "EXP_RECE_UPDATE_ORDER"
	ProductDbMap           = "ProductDbMap"
	ProductQuantityInfoDB  = "ProductQuantityInfoDB"
	QuantityInfoDBList     = "QuantityInfoDBList"
	TemplateCode           = "fm_76130_standard_ZYYMKIJW69LH"
	SfPrice                = "SfPrice"
)

type ServiceContext struct {
	Config            config.Config
	UserShopping      cachemodel.UserShoppingCartModel
	Product           cachemodel.ProductModel
	UserOrder         cachemodel.UserOrderModel
	LocalCache        *collection.Cache
	LocalCache2       *collection.Cache
	UserAddressString cachemodel.UserAddressStringModel
	AccountOperateLog cachemodel.AccountOperateLogModel
	CashAccount       cachemodel.CashAccountModel
	CashLog           cachemodel.CashLogModel
	Client            *core.Client
	FileClient        *cos.Client
	MchPrivateKey     *rsa.PrivateKey
	Handler           *notify.Handler
	Coupon            cachemodel.CouponModel
	UserCoupon        cachemodel.UserCouponModel
	UserPoints        cachemodel.UserPointsModel
	RechargeProduct   cachemodel.RechargeProductModel
	RechargeOrder     cachemodel.RechargeOrderModel
	Invoice           cachemodel.InvoiceModel
	StarMallLongList  cachemodel.StarmallLonglistModel
	TransactionInfo   cachemodel.TransactionInfoModel
	UserInvoiceString cachemodel.UserInvoiceStringModel
	Voucher           cachemodel.VoucherModel
	PointLog          cachemodel.PointLogModel
	Mission           cachemodel.MissionModel
	UserMission       cachemodel.UserMissionModel
	PayInfo           cachemodel.PayInfoModel
	Order             cachemodel.OrderModel
	RefundInfo        cachemodel.RefundInfoModel
	SfInfo            cachemodel.SfInfoModel
	ErrLog            cachemodel.ErrLogModel
	Configuration     cachemodel.ConfigurationModel
	SfPrice           cachemodel.SfPriceModel
	AccessToken       cachemodel.AccessTokenModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	localCache, err := collection.NewCache(localCacheExpire)
	if err != nil {
		panic(err)
	}
	localCache2, err := collection.NewCache(localCacheExpire2)
	if err != nil {
		panic(err)
	}
	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath("etc/apiclient_key.pem")
	//mchPrivateKey, err := utils.LoadPrivateKeyWithPath("C:\\Users\\17854\\Downloads\\host\\apiclient_key.pem")
	if err != nil {
		log.Fatal(err)
	}
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(c.WxConf.MchID, c.WxConf.MchCertificateSerialNumber, mchPrivateKey, c.WxConf.MchAPIv3Key),
		option.WithWechatPayCipher(
			encryptors.NewWechatPayEncryptor(downloader.MgrInstance().GetCertificateVisitor(c.WxConf.MchID)),
			decryptors.NewWechatPayDecryptor(mchPrivateKey),
		),
	}
	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		log.Fatalf("new wechat pay client err:%s", err)
	}
	ctx := context.Background()
	// 1. 使用 `RegisterDownloaderWithPrivateKey` 注册下载器
	err = downloader.MgrInstance().RegisterDownloaderWithPrivateKey(ctx, mchPrivateKey, c.WxConf.MchCertificateSerialNumber, c.WxConf.MchID, c.WxConf.MchAPIv3Key)
	if err != nil {
		log.Fatalf("new wechat verify hadler err:%s", err)
	}
	// 2. 获取商户号对应的微信支付平台证书访问器
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(c.WxConf.MchID)
	// 3. 使用证书访问器初始化 `notify.Handler`
	handler := notify.NewNotifyHandler(c.WxConf.MchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	u, _ := url.Parse("https://kunlun-1310629238.cos.ap-shanghai.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}
	CDNclient := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  "AKIDkKeAQRBftrZ6NN7bpmTu50f4D6i5C2cw", // 用户的 SecretId，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考 https://cloud.tencent.com/document/product/598/37140
			SecretKey: "6vUnqdARSCgtqWw5yby5vWquhFXYJr9B",     // 用户的 SecretKey，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考 https://cloud.tencent.com/document/product/598/37140
		},
	})
	return &ServiceContext{
		Config:            c,
		UserShopping:      cachemodel.NewUserShoppingCartModel(sqlx.NewMysql(c.DB.DataSource)),
		Product:           cachemodel.NewProductModel(sqlx.NewMysql(c.DB.DataSource)),
		LocalCache:        localCache,
		LocalCache2:       localCache2,
		UserAddressString: cachemodel.NewUserAddressStringModel(sqlx.NewMysql(c.DB.DataSource)),
		Client:            client,
		MchPrivateKey:     mchPrivateKey,
		Handler:           handler,
		UserOrder:         cachemodel.NewUserOrderModel(sqlx.NewMysql(c.DB.DataSource)),
		AccountOperateLog: cachemodel.NewAccountOperateLogModel(sqlx.NewMysql(c.DB.DataSource)),
		Coupon:            cachemodel.NewCouponModel(sqlx.NewMysql(c.DB.DataSource)),
		CashAccount:       cachemodel.NewCashAccountModel(sqlx.NewMysql(c.DB.DataSource)),
		UserCoupon:        cachemodel.NewUserCouponModel(sqlx.NewMysql(c.DB.DataSource)),
		CashLog:           cachemodel.NewCashLogModel(sqlx.NewMysql(c.DB.DataSource)),
		UserPoints:        cachemodel.NewUserPointsModel(sqlx.NewMysql(c.DB.DataSource)),
		RechargeProduct:   cachemodel.NewRechargeProductModel(sqlx.NewMysql(c.DB.DataSource)),
		RechargeOrder:     cachemodel.NewRechargeOrderModel(sqlx.NewMysql(c.DB.DataSource)),
		Invoice:           cachemodel.NewInvoiceModel(sqlx.NewMysql(c.DB.DataSource)),
		StarMallLongList:  cachemodel.NewStarmallLonglistModel(sqlx.NewMysql(c.DB.DataSource)),
		TransactionInfo:   cachemodel.NewTransactionInfoModel(sqlx.NewMysql(c.DB.DataSource)),
		UserInvoiceString: cachemodel.NewUserInvoiceStringModel(sqlx.NewMysql(c.DB.DataSource)),
		Voucher:           cachemodel.NewVoucherModel(sqlx.NewMysql(c.DB.DataSource)),
		PointLog:          cachemodel.NewPointLogModel(sqlx.NewMysql(c.DB.DataSource)),
		Mission:           cachemodel.NewMissionModel(sqlx.NewMysql(c.DB.DataSource)),
		UserMission:       cachemodel.NewUserMissionModel(sqlx.NewMysql(c.DB.DataSource)),
		PayInfo:           cachemodel.NewPayInfoModel(sqlx.NewMysql(c.DB.DataSource)),
		Order:             cachemodel.NewOrderModel(sqlx.NewMysql(c.DB.DataSource)),
		RefundInfo:        cachemodel.NewRefundInfoModel(sqlx.NewMysql(c.DB.DataSource)),
		SfInfo:            cachemodel.NewSfInfoModel(sqlx.NewMysql(c.DB.DataSource)),
		ErrLog:            cachemodel.NewErrLogModel(sqlx.NewMysql(c.DB.DataSource)),
		Configuration:     cachemodel.NewConfigurationModel(sqlx.NewMysql(c.DB.DataSource)),
		SfPrice:           cachemodel.NewSfPriceModel(sqlx.NewMysql(c.DB.DataSource)),
		AccessToken:       cachemodel.NewAccessTokenModel(sqlx.NewMysql(c.DB.DataSource)),
		FileClient:        CDNclient,
	}
}
