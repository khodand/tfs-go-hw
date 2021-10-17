package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"time"
)

const (
	cookieAuth           = "auth"
	userID     cookieVal = "ID"
)

type cookieVal string

type Messenger struct {
	messages     []byte
	userMessages map[string][]byte
}

type User struct {
	Login string
}

func NewMessenger() Messenger {
	var m Messenger
	m.userMessages = make(map[string][]byte)

	return m
}

func (m *Messenger) PostMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(cookieVal)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	message := chi.URLParam(r, "msg")

	datetime := time.Now().Format("2006.01.02 15:04:05")
	formatMessage := fmt.Sprintf("{%s %s} %s\n", datetime, id, message)
	m.messages = append(m.messages, []byte(formatMessage)...)
}

func (m *Messenger) PostPrivateMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(cookieVal)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := chi.URLParam(r, "id")
	message := chi.URLParam(r, "msg")

	datetime := time.Now().Format("2006.01.02 15:04:05")
	formatMessage := fmt.Sprintf("{%s %s} %s\n", datetime, id, message)
	if _, ok := m.userMessages[user]; !ok {
		_, _ = w.Write([]byte("No such user :("))
	} else {
		m.userMessages[user] = append(m.userMessages[user], []byte(formatMessage)...)
	}
}

func (m *Messenger) GetMessages(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write(m.messages)
}

func (m *Messenger) GetPrivateMessages(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(cookieVal)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(m.userMessages[string(id)])
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

	if _, ok := m.userMessages[u.Login]; !ok {
		m.userMessages[u.Login] = []byte("Welcome to the chat!!! \n")
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
		idCtx := context.WithValue(r.Context(), userID, cookieVal(c.Value))

		handler.ServeHTTP(w, r.WithContext(idCtx))
	}

	return http.HandlerFunc(fn)
}
