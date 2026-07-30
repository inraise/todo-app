package web_http

import (
	"github.com/inraise/todo-app/internal/core/transport/http/server"
)

type WebHTTPHandler struct {
	webService WebService
}

type WebService interface {
	GetMainPage() ([]byte, error)
}

func NewWebHTTPHandler(
	webService WebService,
) *WebHTTPHandler {
	return &WebHTTPHandler{
		webService: webService,
	}
}

func (h *WebHTTPHandler) Routes() []server.Route {
	return []server.Route {
		{
			Path: "/",
			Handler: h.GetMainPage,
		},
	}
}