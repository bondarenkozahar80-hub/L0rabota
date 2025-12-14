package repository
import (
	"context"
	"order-service0/internal/domain/entities"
)

// OrderRepository определяет контракт для работы с хранилищем заказов
type OrderRepository interface {
	// Create сохраняет заказ в базу данных
	Create(ctx context.Context, order *entities.Order) error
	// GetByUID возвращает заказ по его уникальному идентификатору
	GetByUID(ctx context.Context, orderUID string) (*entities.Order, error)
	// GetAll возвращает все заказы из базы данных
	GetAll(ctx context.Context) ([]*entities.Order, error)
}

// Cache определяет контракт для кэширования заказов в памяти
type Cache interface {
	// Set сохраняет заказ в кэше
	Set(orderUID string, order *entities.Order)
	// Get возвращает заказ из кэша по orderUID
	Get(orderUID string) (*entities.Order, bool)
	// GetAll возвращает все заказы из кэша
	GetAll() map[string]*entities.Order
	// Restore восстанавливает кэш из переданной мапы заказов
	Restore(orders map[string]*entities.Order)
}
