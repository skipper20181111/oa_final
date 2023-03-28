package invoice

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/invoice"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func GetallinvoiceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAllInvoiceRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := invoice.NewGetallinvoiceLogic(r.Context(), svcCtx)
		resp, err := l.Getallinvoice(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
