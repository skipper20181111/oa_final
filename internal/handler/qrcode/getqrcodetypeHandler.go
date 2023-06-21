package qrcode

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/qrcode"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func GetqrcodetypeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ScanQRcodeRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := qrcode.NewGetqrcodetypeLogic(r.Context(), svcCtx)
		resp, err := l.Getqrcodetype(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
