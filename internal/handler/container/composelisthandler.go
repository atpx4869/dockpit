package container

import (
	"net/http"

	"github.com/atpx4869/dockpit/internal/logic/container"
	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ComposeListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := container.NewComposeListLogic(r.Context(), svcCtx)
		resp, err := l.ComposeList()
		if err != nil {
			httpx.WriteJson(w, resp.Code, resp)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
