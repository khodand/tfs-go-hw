package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	m := NewMessenger()

	root := chi.NewRouter()
	root.Use(middleware.Logger)
	root.Post("/login", m.Login)
	root.Get("/messages", m.GetMessages)

	r := chi.NewRouter()
	r.Use(Auth)
	r.Post("/messages/{msg}", m.PostMessage)
	r.Post("/{id}/messages/{msg}", m.PostPrivateMessage)
	r.Get("/me/messages", m.GetPrivateMessages)

	root.Mount("/user", r)

	log.Fatal(http.ListenAndServe(":5000", root))
}
