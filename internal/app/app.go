package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"effective-mobile-subscription-server/internal/config"
	"effective-mobile-subscription-server/internal/domain/service"
	"effective-mobile-subscription-server/internal/presentation/handlers"
	"effective-mobile-subscription-server/internal/repo"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"effective-mobile-subscription-server/docs"
)

type App struct {
	server *http.Server
	db     repo.IDatasource
}

func NewApp(ctx context.Context, cfg *config.Config, logger *logrus.Logger) (*App, error) {
	docs.SwaggerInfo.Title = "Subscription manager"
	docs.SwaggerInfo.Description = "application for managing subscriptions"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "example.com"
	docs.SwaggerInfo.BasePath = "/subscription"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	db, err := repo.NewDatasource(ctx, cfg.GetConnString(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	subService := service.NewService(db, logger)
	handler := handlers.NewHandler(subService, logger)
	router := gin.Default()
	registerHandlers(router, handler)

	server := &http.Server{
		Addr:              cfg.GetAddress(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		Handler:           router,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	return &App{server: server, db: db}, nil
}

func (app *App) Start() error {
	return app.server.ListenAndServe()
}

func (app *App) Stop(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-gCtx.Done()
		return app.server.Shutdown(ctx)
	})
	g.Go(func() error {
		<-gCtx.Done()
		return app.db.Stop(ctx)
	})

	return g.Wait()
}

func registerHandlers(router *gin.Engine, handler handlers.IHandler) {
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))

	subscriptionPath := router.Group(`/subscription`)
	subscriptionPath.GET("/", handler.GetSubscription)
	subscriptionPath.POST("/", handler.AddSubscription)
	subscriptionPath.PUT("/:id", handler.UpdateSubscription)
	subscriptionPath.DELETE("/:id", handler.DeleteSubscription)
	subscriptionPath.GET("/aggregate", handler.GetSubscriptionAggregate)
}
