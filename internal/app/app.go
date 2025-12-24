package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"order-service0/internal/config"
	httpDelivery "order-service0/internal/delivery/http"
	kafkaDelivery "order-service0/internal/delivery/kafka"
	"order-service0/internal/domain/entities"
	"order-service0/internal/repository/cache"
	"order-service0/internal/repository/postgres"
	"order-service0/internal/usecase"

	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type App struct {
	config        *config.Config
	httpServer    *http.Server
	kafkaConsumer *kafkaDelivery.OrderConsumer
	db            *sql.DB
}

func NewApp(cfg *config.Config) *App {
	return &App{config: cfg}
}

func (a *App) initDB() error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		a.config.Database.Host, a.config.Database.Port, a.config.Database.User,
		a.config.Database.Password, a.config.Database.DBName, a.config.Database.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	a.db = db
	return nil
}

func (a *App) initServices() (*httpDelivery.OrderHandler, error) {
	orderRepo := postgres.NewOrderRepository(a.db)
	cacheRepo := cache.NewInMemoryCache()

	ctx := context.Background()
	orders, err := orderRepo.GetAll(ctx)
	if err != nil {
		log.Printf("Warning: Failed to restore cache from database: %v", err)
	} else {
		cacheMap := make(map[string]*entities.Order)
		for _, order := range orders {
			cacheMap[order.OrderUID] = order
		}
		cacheRepo.Restore(cacheMap)
		log.Printf("Restored %d orders to cache", len(orders))
	}

	orderUseCase := usecase.NewOrderUseCase(orderRepo, cacheRepo)

	a.kafkaConsumer = kafkaDelivery.NewOrderConsumer(
		a.config.Kafka.Brokers,
		a.config.Kafka.Topic,
		a.config.Kafka.GroupID,
		orderUseCase,
	)

	return httpDelivery.NewOrderHandler(orderUseCase), nil
}

func (a *App) initHTTPServer(orderHandler *httpDelivery.OrderHandler) {
	router := mux.NewRouter()
	router.HandleFunc("/order/{id}", orderHandler.GetOrderByUID).Methods("GET")
	router.HandleFunc("/", orderHandler.ServeStatic).Methods("GET")
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))),
	)

	a.httpServer = &http.Server{
		Addr:         ":" + a.config.HTTP.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(a.config.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.config.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(a.config.HTTP.IdleTimeout) * time.Second,
	}
}

func (a *App) Run() error {
	if err := a.initDB(); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer a.db.Close()

	orderHandler, err := a.initServices()
	if err != nil {
		return fmt.Errorf("failed to init services: %w", err)
	}

	a.initHTTPServer(orderHandler)

	ctx := context.Background()
	go a.kafkaConsumer.Start(ctx)

	log.Printf("Server starting on port %s", a.config.HTTP.Port)
	if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}

	if a.kafkaConsumer != nil {
		if err := a.kafkaConsumer.Close(); err != nil {
			log.Printf("Kafka consumer close error: %v", err)
		}
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Printf("Database close error: %v", err)
		}
	}
}
