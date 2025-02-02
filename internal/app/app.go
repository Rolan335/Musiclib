package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Rolan335/Musiclib/internal/config"
	"github.com/Rolan335/Musiclib/internal/controller"
	"github.com/Rolan335/Musiclib/internal/logger"
	"github.com/Rolan335/Musiclib/pkg/api"
)

type Service struct {
	server *http.Server
	log    *logger.Log
}

// for graceful shutdown of services (at our case postgres), they should have method Close
type Close interface {
	Close()
}

func NewService(config *config.Config, server *controller.Server, log *logger.Log) *Service {
	gin.SetMode(config.GinMode)
	r := gin.Default()

	r.StaticFile("/openapi.yaml", "./api/musiclib/openapi.yaml")
	r.LoadHTMLGlob("templates/*")
	r.GET("/swagger", func(c *gin.Context) {
		c.HTML(200, "swagger.html", nil)
	})

	api.RegisterHandlers(r, server)

	return &Service{
		server: &http.Server{
			Addr:    config.Port,
			Handler: r,
		},
		log: log,
	}
}

func (s *Service) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			s.log.Logger.Error("server is off", "error", err.Error())
		}
	}()
}

func (s *Service) GracefulStop(services ...interface{}) {
	for _, service := range services {
		if asserted, ok := service.(Close); ok {
			asserted.Close()
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		s.log.Logger.Error("Failed to graceful shutdown", "error", err.Error())
	}
	s.log.Logger.Info("gracefully shut")
}
