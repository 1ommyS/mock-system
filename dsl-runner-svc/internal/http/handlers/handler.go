package handlers

import "dsl-runner-svc/internal/application"

type Handler struct {
	Service *application.Service
}

func New(service *application.Service) *Handler {
	return &Handler{Service: service}
}
