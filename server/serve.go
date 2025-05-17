package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lengzuo/fundflow/configs"
	"github.com/lengzuo/fundflow/dao"
	"github.com/lengzuo/fundflow/pkg/log"
	pkgredis "github.com/lengzuo/fundflow/pkg/redis"
	"github.com/lengzuo/fundflow/server/middlewares"
	"github.com/lengzuo/fundflow/usecases/users"
	"github.com/lengzuo/fundflow/utils"
	"github.com/redis/go-redis/v9"
)

func Serve() {
	config, err := configs.New()
	if err != nil {
		panic(fmt.Sprintf("failed in loading config with err: %s", err))
	}
	log.New(config.Mode)

	serverCtx, serverStopCtx := context.WithTimeout(context.Background(), 60*time.Second)
	defer serverStopCtx()

	// Initialize Redis client
	redisClient := pkgredis.New(config.RedisConfig.URL)

	// Initialize Database client
	db, err := dao.New(serverCtx, config.DatabaseConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// Initialize DAOs from database client above
	userDAO := dao.NewUsers(db)
	_ = dao.NewTransactions(db)
	_ = dao.NewWallets(db)
	_ = dao.NewLedgers(db)

	// Initialize usecases
	userServices := users.New(userDAO)

	// The HTTP Server
	server := &http.Server{
		Addr: "0.0.0.0:8080",
		Handler: router(
			redisClient,
			userServices,
		),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
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

func router(
	redisClient *redis.Client,
	userServices users.Service,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(utils.APIRequestTimeout))
	r.Use(middlewares.Idempotency(redisClient))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	r.Get("/long", func(w http.ResponseWriter, r *http.Request) {
		log.Debug(r.Context(), "long task")
		time.Sleep(1 * time.Minute)
		w.Write([]byte("hi"))
	})

	r.Route("/api", func(apiRouter chi.Router) {
		// No Auth API
		apiRouter.Mount("/public/users", usersRouter(userServices))
		// Auth API

	})

	return r
}
