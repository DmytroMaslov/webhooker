package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"webhooker/api"
	"webhooker/api/handlers"
	"webhooker/config"
	"webhooker/internal/services"
	"webhooker/internal/storage/posgres"
)

type App struct {
	Config *config.Config
}

func (a *App) Run() {
	log.Printf("App started\n")

	eventStorage, err := posgres.NewEventStorage(&a.Config.Postgress)
	if err != nil {
		log.Fatal("failed to create event storage: %w", err)
	}
	orderStorage, err := posgres.NewOrderStorage(&a.Config.Postgress)
	if err != nil {
		log.Fatal("failed to create event storage: %w", err)
	}

	streamService := services.NewStreamService(eventStorage, orderStorage)

	webhookHandler := handlers.NewHandler(streamService)

	server := api.NewHttpServer(8080, webhookHandler.GetHandlers())

	err = server.Serve()
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to serve, err: %s", err))
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	<-exit

	// shout down logic
	err = server.Shutdown(context.Background())
	if err != nil {
		log.Printf("failed to stop server, err: %s", err)
	}

	err = eventStorage.Close()
	if err != nil {
		log.Printf("failed to close connection, err: %s", err)
	}
	log.Printf("See you\n")
}
