package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func execute[Resp any](w http.ResponseWriter, r *http.Request, call func() (*Resp, error)) {
	resp, err := call()
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func parseAndExecute[Req any, Resp any](w http.ResponseWriter, r *http.Request, call func(*Req) (*Resp, error)) {
	var req Req
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	execute(w, r, func() (*Resp, error) {
		return call(&req)
	})
}
