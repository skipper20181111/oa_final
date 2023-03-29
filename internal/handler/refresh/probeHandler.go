package refresh

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/refresh"
	"oa_final/internal/svc"
)

func ProbeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := refresh.NewProbeLogic(r.Context(), svcCtx)
		err := l.Probe()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
