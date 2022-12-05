package refresh

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/refresh"
	"oa_final/internal/svc"
)

func RefreshPLHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := refresh.NewRefreshPLLogic(r.Context(), svcCtx)
		resp, err := l.RefreshPL()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
