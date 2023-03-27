package payrecall

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/payrecall"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func CoupontellmesoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TellMeSoRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := payrecall.NewCoupontellmesoLogic(r.Context(), svcCtx)
		resp, err := l.Coupontellmeso(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
