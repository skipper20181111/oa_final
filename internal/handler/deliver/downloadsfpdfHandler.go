package deliver

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/deliver"
	"oa_final/internal/svc"
)

func DownloadsfpdfHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := deliver.NewDownloadsfpdfLogic(context.Background(), svcCtx)
		resp, err := l.Downloadsfpdf()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
