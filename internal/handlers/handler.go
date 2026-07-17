package handlers

import (
	"URLShortener/internal/service"
)

type Handler struct {
	service service.URLService
}

func New(service service.URLService) *Handler {
	return &Handler{
		service: service,
	}
}
