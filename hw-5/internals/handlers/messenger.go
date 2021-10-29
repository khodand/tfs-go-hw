package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"tfs-go-hw/hw-5/internals/services"
)

type MessengerService interface {
	PostMessage(w http.ResponseWriter, r *http.Request)
	PostPrivateMessage(w http.ResponseWriter, r *http.Request)
	GetMessages(w http.ResponseWriter, r *http.Request)
	GetPrivateMessages(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

type Messenger struct {
	service MessengerService
}

func NewMessenger(messenger MessengerService) *Messenger {
	return &Messenger{
		service: messenger,
	}
}

func (m *Messenger) Router() chi.Router {
	root := chi.NewRouter()
	root.Use(middleware.Logger)
	root.Post("/login", m.service.Login)
	root.Get("/messages", m.service.GetMessages)

	r := chi.NewRouter()
	r.Use(services.Auth)
	r.Post("/messages/{msg}", m.service.PostMessage)
	r.Post("/{id}/messages/{msg}", m.service.PostPrivateMessage)
	r.Get("/me/messages", m.service.GetPrivateMessages)

	root.Mount("/user", r)

	return root
}
