package services

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"tfs-go-hw/hw-5/internals/domain"
)

const (
	cookieAuth                  = "auth"
	userID     domain.CookieVal = "ID"
)

type User struct {
	Login string
}

type MessengerDatabase interface {
	AddMessage(message string, id domain.CookieVal)
	AddUserMessage(message string, user domain.CookieVal, receiverID string) error
	GetMessages() []byte
	GetPrivateMessages(id domain.CookieVal) ([]byte, error)
	CreateNewUser(id domain.CookieVal)
}

type Messenger struct {
	database MessengerDatabase
}

func NewMessenger(db MessengerDatabase) *Messenger {
	return &Messenger{
		database: db,
	}
}

func (m *Messenger) PostMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(domain.CookieVal)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m.database.AddMessage(chi.URLParam(r, "msg"), id)
}

func (m *Messenger) PostPrivateMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(domain.CookieVal)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := chi.URLParam(r, "id")
	message := chi.URLParam(r, "msg")
	if err := m.database.AddUserMessage(message, id, user); err != nil {
		_, _ = w.Write([]byte(err.Error()))
	}
}

func (m *Messenger) GetMessages(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write(m.database.GetMessages())
}

func (m *Messenger) GetPrivateMessages(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(domain.CookieVal)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	messages, err := m.database.GetPrivateMessages(id)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
	}
	_, _ = w.Write(messages)
}

func (m *Messenger) Login(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var u User
	err = json.Unmarshal(d, &u)
	if err != nil || u.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c := &http.Cookie{
		Name:  cookieAuth,
		Value: u.Login,
		Path:  "/",
	}

	if _, err := m.database.GetPrivateMessages(domain.CookieVal(u.Login)); err != nil {
		m.database.CreateNewUser(domain.CookieVal(u.Login))
	}

	http.SetCookie(w, c)
}

func Auth(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieAuth)
		switch err {
		case nil:
		case http.ErrNoCookie:
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if c.Value == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		idCtx := context.WithValue(r.Context(), userID, domain.CookieVal(c.Value))

		handler.ServeHTTP(w, r.WithContext(idCtx))
	}

	return http.HandlerFunc(fn)
}
