package orderpay

import (
	"context"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/orderpay"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func RefundallorderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CancelAllOrderRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		NewCtx, _ := context.WithTimeout(r.Context(), time.Second*20)
		l := orderpay.NewRefundallorderLogic(NewCtx, svcCtx)
		resp, err := l.Refundallorder(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
