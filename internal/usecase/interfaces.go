package usecase

import (
	"context"
	"order-service0/internal/domain/entities"
)

// OrderUseCase определяет бизнес-логику работы с заказами
type OrderUseCase interface {
	CreateOrder(ctx context.Context, order *entities.Order) error
	GetOrderByUID(ctx context.Context, orderUID string) (*entities.Order, error)
	ProcessOrderMessage(ctx context.Context, message []byte) error
}

// OrderRepository определяет контракт для работы с хранилищем заказов
type OrderRepository interface {
	Create(ctx context.Context, order *entities.Order) error
	GetByUID(ctx context.Context, orderUID string) (*entities.Order, error)
	GetAll(ctx context.Context) ([]*entities.Order, error)
}

// Cache определяет контракт для кэширования
type Cache interface {
	Set(orderUID string, order *entities.Order)
	Get(orderUID string) (*entities.Order, bool)
	GetAll() map[string]*entities.Order
	Restore(orders map[string]*entities.Order)
}
