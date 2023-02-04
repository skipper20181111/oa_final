package cachemodel

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserCouponModel = (*customUserCouponModel)(nil)

type (
	// UserCouponModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserCouponModel.
	UserCouponModel interface {
		userCouponModel
	}

	customUserCouponModel struct {
		*defaultUserCouponModel
	}
)

// NewUserCouponModel returns a model for the database table.
func NewUserCouponModel(conn sqlx.SqlConn, c cache.CacheConf) UserCouponModel {
	return &customUserCouponModel{
		defaultUserCouponModel: newUserCouponModel(conn, c),
	}
}
