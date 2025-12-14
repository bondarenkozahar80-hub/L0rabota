package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"order-service0/internal/app"    // добавить
	"order-service0/internal/config" // добавить
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создание приложения
	application := app.NewApp(cfg)

	// Обработка сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск приложения в горутине
	go func() {
		if err := application.Run(); err != nil {
			log.Fatalf("Application run error: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down...", sig)

	// Graceful shutdown
	application.Stop()
	log.Println("Application stopped")
}
