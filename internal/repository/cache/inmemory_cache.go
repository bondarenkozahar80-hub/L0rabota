package cache

import (

	"sync"

	"order-service0/internal/domain/entities"
)

type inMemoryCache struct {
	mu     sync.RWMutex
	orders map[string]*entities.Order
}

func NewInMemoryCache() *inMemoryCache {
	return &inMemoryCache{
		orders: make(map[string]*entities.Order),
	}
}

func (c *inMemoryCache) Set(orderUID string, order *entities.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[orderUID] = order
}

func (c *inMemoryCache) Get(orderUID string) (*entities.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.orders[orderUID]
	return order, exists
}

func (c *inMemoryCache) GetAll() map[string]*entities.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]*entities.Order)
	for k, v := range c.orders {
		result[k] = v
	}
	return result
}

func (c *inMemoryCache) Restore(orders map[string]*entities.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders = orders
}
