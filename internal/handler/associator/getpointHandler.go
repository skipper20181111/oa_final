package associator

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/associator"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func GetpointHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPointRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := associator.NewGetpointLogic(r.Context(), svcCtx)
		resp, err := l.Getpoint(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
