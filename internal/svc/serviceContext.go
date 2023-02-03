package svc

import (
	"context"
	"crypto/rsa"
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
	"oa_final/cachemodel"
	"oa_final/internal/config"
	"time"
)

const localCacheExpire = time.Duration(time.Second * 800)

type ServiceContext struct {
	Config            config.Config
	UserShopping      cachemodel.UserShoppingCartModel
	Product           cachemodel.ProductModel
	UserOrder         cachemodel.UserOrderModel
	LocalCache        *collection.Cache
	UserAddressString cachemodel.UserAddressStringModel
	AccountOperateLog cachemodel.AccountOperateLogModel
	CashAccount       cachemodel.CashAccountModel
	Client            *core.Client
	MchPrivateKey     *rsa.PrivateKey
	Handler           *notify.Handler
	Coupon            cachemodel.CouponModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	localCache, err := collection.NewCache(localCacheExpire)
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

	return &ServiceContext{
		Config:            c,
		UserShopping:      cachemodel.NewUserShoppingCartModel(sqlx.NewMysql(c.DB.DataSource), c.Cache),
		Product:           cachemodel.NewProductModel(sqlx.NewMysql(c.DB.DataSource)),
		LocalCache:        localCache,
		UserAddressString: cachemodel.NewUserAddressStringModel(sqlx.NewMysql(c.DB.DataSource)),
		Client:            client,
		MchPrivateKey:     mchPrivateKey,
		Handler:           handler,
		UserOrder:         cachemodel.NewUserOrderModel(sqlx.NewMysql(c.DB.DataSource)),
		AccountOperateLog: cachemodel.NewAccountOperateLogModel(sqlx.NewMysql(c.DB.DataSource)),
		Coupon:            cachemodel.NewCouponModel(sqlx.NewMysql(c.DB.DataSource), c.Cache),
		CashAccount:       cachemodel.NewCashAccountModel(sqlx.NewMysql(c.DB.DataSource), c.Cache),
	}
}
