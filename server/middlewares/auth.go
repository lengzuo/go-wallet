package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/lengzuo/fundflow/internal/apierr"
	"github.com/lengzuo/fundflow/pkg/log"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authUsername := r.Header.Get("Authorization")
		if strings.TrimSpace(authUsername) == "" {
			err := apierr.Unauthenticated()
			render.Status(r, err.HTTPStatusCode())
			render.JSON(w, r, err)
			return
		}
		r = r.WithContext(context.WithValue(ctx, log.UsernameKey, authUsername))
		next.ServeHTTP(w, r)
	})
}
