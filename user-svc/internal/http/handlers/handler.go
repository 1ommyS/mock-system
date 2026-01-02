package handlers

import "user-svc/internal/application"

type Handler struct {
	Services *application.Service
}

func New(services *application.Service) *Handler {
	return &Handler{
		Services: services,
	}
}
