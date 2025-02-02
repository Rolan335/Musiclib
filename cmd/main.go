//go:generate oapi-codegen -generate types,models,gin -package api -o ../pkg/api/api.gen.go ../api/swagger/openapi.yaml
//go:generate oapi-codegen -generate client,models,types -package musicinfo -o ../pkg/musicinfo/api.gen.go ../api/external/musicinfo.yaml
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rolan335/Musiclib/internal/app"
	"github.com/Rolan335/Musiclib/internal/config"
	"github.com/Rolan335/Musiclib/internal/controller"
	"github.com/Rolan335/Musiclib/internal/logger"
	"github.com/Rolan335/Musiclib/internal/musiclib"
	"github.com/Rolan335/Musiclib/internal/repository/postgres"
)

// @title Music Library API
// @version 1.0
// @description Music Library API documentation
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

func main() {
	//load config from env
	cfg := config.MustNewConfig()

	//Initializing logger
	out := os.Stdout
	logger := logger.New(cfg.LogLevel, out)

	//making migration
	if err := postgres.Migrate(&cfg.Migration); err != nil {
		panic("failed to do migrations: " + err.Error())
	}

	//Initializing postgres storage
	storage := postgres.MustNewStorage(&cfg.DB, logger)

	//Initializing business logic
	musiclib := musiclib.NewMusicLib(storage, logger)

	//creating server controller with handlers
	server := controller.MustNewServer(musiclib, cfg.API.URL, cfg.RequestTimeout)

	//starting http service
	app := app.NewService(cfg, server, logger)

	//creating notify ctx for graceful shutdown and starting app
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	app.MustStart()
	<-ctx.Done()

	//stopping server and provided services. Provided servies should have method Stop()
	app.GracefulStop(storage)
}
