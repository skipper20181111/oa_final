package svc

import (
	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"oa_final/cachemodel"
	"oa_final/internal/config"

	"time"
)

const localCacheExpire = time.Duration(time.Second * 800)

type ServiceContext struct {
	Config            config.Config
	UserShopping      cachemodel.UserShoppingCartModel
	Product           cachemodel.ProductModel
	LocalCache        *collection.Cache
	UserAddressString cachemodel.UserAddressStringModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	localCache, err := collection.NewCache(localCacheExpire)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:            c,
		UserShopping:      cachemodel.NewUserShoppingCartModel(sqlx.NewMysql(c.DB.DataSource), c.Cache),
		Product:           cachemodel.NewProductModel(sqlx.NewMysql(c.DB.DataSource), c.Cache),
		LocalCache:        localCache,
		UserAddressString: cachemodel.NewUserAddressStringModel(sqlx.NewMysql(c.DB.DataSource), c.Cache),
	}
}
