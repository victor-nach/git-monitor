package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/victor-nach/git-monitor/internal/config"
	"github.com/victor-nach/git-monitor/internal/http/handlers"
	"go.uber.org/zap"
)

func Run(log *zap.Logger, handler *handlers.Handler, cfg *config.Config) {
	log = log.With(zap.String("service", "http-server"))

	log.Info("Starting http server", zap.String("address", cfg.Port))

	router := gin.Default()

	// Set up CORS middleware
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type"},
		ExposeHeaders:   []string{"Content-Length"},
		MaxAge:          12 * time.Hour,
	}))

	router.GET("/", welcomeHandler)

	api := router.Group("/api/v1")
	{

		api.GET("/tasks", handler.ListTasks)
		api.GET("/tasks/:id", handler.GetTask)

		repos := api.Group("/repos")
		{
			repos.GET("/", handler.ListTrackedRepositories)

			repo := repos.Group("/:owner/:repo")
			repo.Use(handler.RepoInfoMiddleware)
			{
				repo.POST("", handler.AddTrackedRepository)
				repo.GET("/top-authors", handler.GetTopCommitAuthors)
				repo.GET("/commits", handler.ListCommits)
				repo.POST("/trigger", handler.TriggerTask)
				repo.PATCH("/status", handler.UpdateRepoStatus)
				repo.POST("/reset", handler.ResetRepo)
			}
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Info("Starting server", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Attempt graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exiting")
}

func welcomeHandler(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Git Monitor API")
}
