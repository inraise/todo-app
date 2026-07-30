package web_http

import (
	"net/http"

	"github.com/inraise/todo-app/internal/core/logger"
	"github.com/inraise/todo-app/internal/core/transport/http/response"
)

func (h* WebHTTPHandler) GetMainPage(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewHTTPResponseHandler(log, rw)

	html, err := h.webService.GetMainPage()
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get index.html for main page",
		)

		return
	}

	responseHandler.HTMLResponse(html)
	 
}	