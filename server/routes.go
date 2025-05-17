package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lengzuo/fundflow/usecases/users"
)

func usersRouter(users users.Service) http.Handler {
	r := chi.NewRouter()
	r.Post("/signup", Handle(users.SignUp))
	return r
}
