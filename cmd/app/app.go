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

	dbClient, err := posgres.NewPgClient(&a.Config.Postgress)
	if err != nil {
		log.Fatal("failed to create db client %w", err)
	}

	eventStorage := posgres.NewEventStorage(dbClient)
	orderStorage := posgres.NewOrderStorage(dbClient)

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

	// shout down logic
	<-exit

	broker.Close()

	err = server.Shutdown(context.Background())
	if err != nil {
		log.Printf("failed to stop server, err: %s", err)
	}

	<-delay.GracefulExit()

	err = dbClient.Close()
	if err != nil {
		log.Printf("failed to close db connection, err: %s", err)
	}
}
