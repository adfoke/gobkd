package app

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	authx "gobkd/internal/auth"
	"gobkd/internal/config"
	"gobkd/internal/handler"
	"gobkd/internal/logger"
	appmw "gobkd/internal/middleware"
	"gobkd/internal/repository"
	"gobkd/internal/response"
	"gobkd/internal/service"
	"gobkd/migrations"
)

func Run(ctx context.Context) error {
	cfg := config.Load()

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return err
	}

	db, err := openDB(cfg.SQLitePath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := runMigrations(ctx, db); err != nil {
		return err
	}

	authService := authx.New(cfg)
	router := buildRouter(db, log, authService)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.WithFields(logrus.Fields{
		"addr":        cfg.HTTPAddr,
		"app_env":     cfg.AppEnv,
		"sqlite_path": cfg.SQLitePath,
	}).Info("server starting")

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == nil || err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func buildRouter(db *sql.DB, log *logrus.Logger, authService *authx.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(appmw.RequestID())
	router.Use(appmw.SecurityHeaders())
	router.Use(appmw.Recovery(log))
	router.Use(appmw.RequestLogger(log))

	systemHandler := handler.NewSystemHandler(db)
	demoHandler := handler.NewDemoHandler()
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(authService, userService)

	router.GET("/ping", systemHandler.Ping)
	router.GET("/healthz", systemHandler.Healthz)
	router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "route not found")
	})
	router.NoMethod(func(c *gin.Context) {
		response.MethodNotAllowed(c, "method not allowed")
	})

	router.Any("/auth/*path", gin.WrapH(http.StripPrefix("/auth", authService.Routes())))

	api := router.Group("/api/v1")
	api.Use(wrapHTTPMiddleware(authService.Trace))
	api.Use(appmw.RequireUser(authService))
	api.GET("/me", userHandler.Me)
	api.POST("/echo", demoHandler.Echo)

	return router
}

func wrapHTTPMiddleware(mw func(http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		calledNext := false

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calledNext = true
			c.Request = r
			c.Next()
		})

		mw(next).ServeHTTP(c.Writer, c.Request)

		if !calledNext {
			c.Abort()
		}
	}
}

func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	entries, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		query, err := fs.ReadFile(migrations.Files, entry.Name())
		if err != nil {
			return err
		}

		if _, err := db.ExecContext(ctx, string(query)); err != nil {
			return err
		}
	}

	return nil
}
