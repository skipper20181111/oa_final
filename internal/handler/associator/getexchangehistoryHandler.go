package associator

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/associator"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func GetexchangehistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetExchangeHistoryRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := associator.NewGetexchangehistoryLogic(r.Context(), svcCtx)
		resp, err := l.Getexchangehistory(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
