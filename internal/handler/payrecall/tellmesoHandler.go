package payrecall

import (
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"

	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"oa_final/internal/logic/payrecall"
	"oa_final/internal/svc"
)

func TellmesoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := payrecall.NewTellmesoLogic(r.Context(), svcCtx)
		transaction := new(payments.Transaction)
		//var transaction *types.SuccessInfo
		notifyReq, err := svcCtx.Handler.ParseNotifyRequest(r.Context(), r, transaction)
		defer notifyReq.RawRequest.Body.Close()
		if err != nil {
			fmt.Println("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&,解密失败了", err.Error(), "&&&&&&&&&&&&&&&&&&&&&&&&")
			httpx.Error(w, err)
		}

		resp, err := l.Tellmeso(notifyReq, transaction)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
