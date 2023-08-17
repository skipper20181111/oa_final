package deliver

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/deliver"
	"oa_final/internal/svc"
)

func StopprepareHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := deliver.NewStopprepareLogic(r.Context(), svcCtx)
		resp, err := l.Stopprepare()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
