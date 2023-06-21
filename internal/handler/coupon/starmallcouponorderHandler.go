package coupon

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/coupon"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func StarmallcouponorderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StarMallCouponOrderRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := coupon.NewStarmallcouponorderLogic(r.Context(), svcCtx)
		resp, err := l.Starmallcouponorder(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
