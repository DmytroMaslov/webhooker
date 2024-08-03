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
	"webhooker/internal/queue/inmemory"
	"webhooker/internal/schedule/delay"
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

	broker := inmemory.NewBroker()

	delay := delay.NewDelay()

	webhookService := services.NewWebhookService(eventStorage, orderStorage, broker, delay)
	orderService := services.NewOrderService(orderStorage)

	handlers := handlers.NewHandler(webhookService, orderService)

	server := api.NewHttpServer(8080, handlers.GetHandlers())

	err = server.Serve()
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to serve, err: %s", err))
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	<-exit

	broker.Close()

	// shout down logic
	err = server.Shutdown(context.Background())
	if err != nil {
		log.Printf("failed to stop server, err: %s", err)
	}

	<-delay.GracefulExit()

	err = eventStorage.Close()
	if err != nil {
		log.Printf("failed to close connection, err: %s", err)
	}

	err = orderStorage.Close()
	if err != nil {
		log.Printf("failed to close connection, err: %s", err)
	}

	log.Printf("See you\n")
}
