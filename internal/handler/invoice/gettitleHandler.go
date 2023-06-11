package invoice

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/invoice"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func GettitleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTitleRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := invoice.NewGettitleLogic(r.Context(), svcCtx)
		resp, err := l.Gettitle(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
