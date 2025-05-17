package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/lengzuo/fundflow/configs"
	"github.com/lengzuo/fundflow/dao"
	"github.com/lengzuo/fundflow/internal/apierr"
	"github.com/lengzuo/fundflow/pkg/log"
)

func Serve() {
	config, err := configs.New()
	if err != nil {
		panic(fmt.Sprintf("failed in loading config with err: %s", err))
	}
	log.New(config.Mode)

	serverCtx, serverStopCtx := context.WithTimeout(context.Background(), 60*time.Second)
	defer serverStopCtx()

	// Initialize Database client
	_, err = dao.New(serverCtx, config.DatabaseConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// The HTTP Server
	server := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      router(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Info(serverCtx, "starting server on %s in [%s] mode", server.Addr, config.Mode.String())
		// Start the server
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(serverCtx, "failed in starting server with err: %s", err)
		}
	}()

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	quit := <-sig

	log.Info(serverCtx, "starting graceful shutdown for server")
	// Server run context
	err = server.Shutdown(serverCtx)
	if err != nil {
		log.Error(serverCtx, "failed in shutdown server: %v", err)
	}
	log.Info(serverCtx, "server shutdown successfully, quit signal: %s", quit.String())
}

func router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	r.Get("/long", func(w http.ResponseWriter, r *http.Request) {
		log.Debug(r.Context(), "long task")
		time.Sleep(1 * time.Minute)
		w.Write([]byte("hi"))
	})

	return r
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authUsername := r.Header.Get("Authorization")
		if strings.TrimSpace(authUsername) == "" {
			err := apierr.Unauthorized()
			render.Status(r, err.HTTPStatusCode())
			render.JSON(w, r, err)
			return
		}
		r = r.WithContext(context.WithValue(ctx, log.UsernameKey, authUsername))
		log.Debug(ctx, "ddrun auth middleware")
		next.ServeHTTP(w, r)
	})
}
