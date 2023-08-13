package deliver

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/deliver"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func DownloadfailedHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DownLoadedRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := deliver.NewDownloadfailedLogic(r.Context(), svcCtx)
		err := l.Downloadfailed(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
