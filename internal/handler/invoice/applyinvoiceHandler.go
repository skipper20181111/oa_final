package invoice

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa_final/internal/logic/invoice"
	"oa_final/internal/svc"
	"oa_final/internal/types"
)

func ApplyinvoiceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApplyInvoiceRes
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := invoice.NewApplyinvoiceLogic(r.Context(), svcCtx)
		resp, err := l.Applyinvoice(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
